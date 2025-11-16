#!/usr/bin/env python3
"""
Dynamic Campaign Data Loader (DCDL) Service
Handles file uploads, database queries, and parameter mapping for campaigns
"""

import logging
import csv
import json
import pandas as pd
from typing import List, Dict, Any, Optional, Tuple
from datetime import datetime, timedelta
from io import StringIO, BytesIO
from sqlalchemy import and_, or_, text
from sqlalchemy.orm import Session

from src.models.profiling import Profile

logger = logging.getLogger(__name__)


class DCDLDataset:
    """Model for DCDL Dataset (matches tbl_dcdl_datasets)"""
    def __init__(self, **kwargs):
        for key, value in kwargs.items():
            setattr(self, key, value)


class DCDLService:
    """Service for managing dynamic campaign data"""

    def create_dataset(
        self,
        db: Session,
        customer_id: int,
        user_id: int,
        dataset_config: Dict[str, Any]
    ) -> Dict[str, Any]:
        """
        Create a new DCDL dataset

        Args:
            db: Database session
            customer_id: Customer ID
            user_id: User ID
            dataset_config: Dataset configuration

        Returns:
            Created dataset info
        """
        try:
            dataset_code = self._generate_dataset_code(db)

            # Insert dataset record
            insert_query = text("""
                INSERT INTO tbl_dcdl_datasets
                (dataset_code, customer_id, user_id, source_type, dataset_name,
                 description, file_name, file_type, file_size_kb, columns,
                 total_records, valid_records, invalid_records, status,
                 created_by, created_at)
                VALUES
                (:dataset_code, :customer_id, :user_id, :source_type, :dataset_name,
                 :description, :file_name, :file_type, :file_size_kb, :columns::jsonb,
                 :total_records, :valid_records, :invalid_records, :status,
                 :created_by, NOW())
                RETURNING dataset_id, dataset_code
            """)

            result = db.execute(insert_query, {
                'dataset_code': dataset_code,
                'customer_id': customer_id,
                'user_id': user_id,
                'source_type': dataset_config.get('source_type', 'FILE_UPLOAD'),
                'dataset_name': dataset_config.get('dataset_name'),
                'description': dataset_config.get('description'),
                'file_name': dataset_config.get('file_name'),
                'file_type': dataset_config.get('file_type'),
                'file_size_kb': dataset_config.get('file_size_kb', 0),
                'columns': json.dumps(dataset_config.get('columns', [])),
                'total_records': 0,
                'valid_records': 0,
                'invalid_records': 0,
                'status': 'PENDING',
                'created_by': str(user_id)
            })

            row = result.fetchone()
            db.commit()

            logger.info(f"Created DCDL dataset {row[1]}")
            return {
                'dataset_id': row[0],
                'dataset_code': row[1],
                'status': 'PENDING'
            }

        except Exception as e:
            db.rollback()
            logger.error(f"Error creating dataset: {str(e)}")
            raise

    def upload_csv_data(
        self,
        db: Session,
        dataset_id: int,
        file_content: bytes,
        column_mapping: Dict[str, str]
    ) -> Dict[str, Any]:
        """
        Upload CSV data to dataset

        Args:
            db: Database session
            dataset_id: Dataset ID
            file_content: CSV file content
            column_mapping: Column to parameter mapping

        Returns:
            Upload statistics
        """
        try:
            # Update dataset status
            db.execute(text("""
                UPDATE tbl_dcdl_datasets
                SET status = 'PROCESSING', processing_started_at = NOW()
                WHERE dataset_id = :dataset_id
            """), {'dataset_id': dataset_id})
            db.commit()

            # Parse CSV
            csv_text = file_content.decode('utf-8')
            csv_reader = csv.DictReader(StringIO(csv_text))

            rows = list(csv_reader)
            total_records = len(rows)
            valid_records = 0
            invalid_records = 0

            # Process and cache rows
            for idx, row in enumerate(rows):
                try:
                    # Map columns
                    mapped_data = {}
                    for source_col, target_param in column_mapping.items():
                        if source_col in row:
                            mapped_data[target_param] = row[source_col]

                    # Extract MSISDN if present
                    msisdn = mapped_data.get('msisdn', mapped_data.get('phone', ''))

                    # Insert into cache
                    db.execute(text("""
                        INSERT INTO tbl_dcdl_data_cache
                        (dataset_id, record_index, record_data, msisdn,
                         is_valid, created_at)
                        VALUES
                        (:dataset_id, :record_index, :record_data::jsonb,
                         :msisdn, TRUE, NOW())
                    """), {
                        'dataset_id': dataset_id,
                        'record_index': idx,
                        'record_data': json.dumps(mapped_data),
                        'msisdn': msisdn
                    })

                    valid_records += 1

                except Exception as row_error:
                    logger.warning(f"Invalid row {idx}: {str(row_error)}")
                    invalid_records += 1

            # Update dataset statistics
            db.execute(text("""
                UPDATE tbl_dcdl_datasets
                SET total_records = :total_records,
                    valid_records = :valid_records,
                    invalid_records = :invalid_records,
                    status = 'COMPLETED',
                    processing_completed_at = NOW(),
                    cache_expiry = NOW() + INTERVAL '7 days'
                WHERE dataset_id = :dataset_id
            """), {
                'dataset_id': dataset_id,
                'total_records': total_records,
                'valid_records': valid_records,
                'invalid_records': invalid_records
            })

            db.commit()

            logger.info(f"Uploaded {valid_records} records to dataset {dataset_id}")
            return {
                'total_records': total_records,
                'valid_records': valid_records,
                'invalid_records': invalid_records,
                'status': 'COMPLETED'
            }

        except Exception as e:
            db.rollback()
            # Mark as failed
            db.execute(text("""
                UPDATE tbl_dcdl_datasets
                SET status = 'FAILED', error_message = :error
                WHERE dataset_id = :dataset_id
            """), {'dataset_id': dataset_id, 'error': str(e)})
            db.commit()

            logger.error(f"Error uploading CSV data: {str(e)}")
            raise

    def create_database_query_dataset(
        self,
        db: Session,
        customer_id: int,
        user_id: int,
        query_config: Dict[str, Any]
    ) -> Dict[str, Any]:
        """
        Create dataset from database query

        Args:
            db: Database session
            customer_id: Customer ID
            user_id: User ID
            query_config: Query configuration

        Returns:
            Dataset info
        """
        try:
            dataset_code = self._generate_dataset_code(db)

            # Insert dataset
            result = db.execute(text("""
                INSERT INTO tbl_dcdl_datasets
                (dataset_code, customer_id, user_id, source_type,
                 dataset_name, description, status, created_by, created_at)
                VALUES
                (:dataset_code, :customer_id, :user_id, 'DATABASE_QUERY',
                 :dataset_name, :description, 'PENDING', :created_by, NOW())
                RETURNING dataset_id, dataset_code
            """), {
                'dataset_code': dataset_code,
                'customer_id': customer_id,
                'user_id': user_id,
                'dataset_name': query_config.get('dataset_name'),
                'description': query_config.get('description'),
                'created_by': str(user_id)
            })

            row = result.fetchone()
            dataset_id = row[0]

            # Insert query definition
            db.execute(text("""
                INSERT INTO tbl_dcdl_queries
                (dataset_id, query_type, query_text, query_parameters,
                 refresh_frequency, is_active, created_at)
                VALUES
                (:dataset_id, :query_type, :query_text, :query_parameters::jsonb,
                 :refresh_frequency, TRUE, NOW())
            """), {
                'dataset_id': dataset_id,
                'query_type': query_config.get('query_type', 'SQL'),
                'query_text': query_config.get('query_text'),
                'query_parameters': json.dumps(query_config.get('parameters', {})),
                'refresh_frequency': query_config.get('refresh_frequency', 'MANUAL')
            })

            db.commit()

            # Execute query immediately
            self.refresh_query_dataset(db, dataset_id)

            logger.info(f"Created query dataset {dataset_code}")
            return {
                'dataset_id': dataset_id,
                'dataset_code': dataset_code,
                'status': 'COMPLETED'
            }

        except Exception as e:
            db.rollback()
            logger.error(f"Error creating query dataset: {str(e)}")
            raise

    def refresh_query_dataset(
        self,
        db: Session,
        dataset_id: int
    ) -> Dict[str, Any]:
        """
        Refresh dataset from database query

        Args:
            db: Database session
            dataset_id: Dataset ID

        Returns:
            Refresh statistics
        """
        try:
            # Get query
            query_result = db.execute(text("""
                SELECT query_text, query_parameters
                FROM tbl_dcdl_queries
                WHERE dataset_id = :dataset_id
                LIMIT 1
            """), {'dataset_id': dataset_id}).fetchone()

            if not query_result:
                raise ValueError(f"No query found for dataset {dataset_id}")

            query_text = query_result[0]
            query_params = json.loads(query_result[1]) if query_result[1] else {}

            # Execute query
            results = db.execute(text(query_text), query_params).fetchall()

            # Clear existing cache
            db.execute(text("""
                DELETE FROM tbl_dcdl_data_cache
                WHERE dataset_id = :dataset_id
            """), {'dataset_id': dataset_id})

            # Insert new data
            valid_records = 0
            for idx, row in enumerate(results):
                record_data = dict(row._mapping)

                db.execute(text("""
                    INSERT INTO tbl_dcdl_data_cache
                    (dataset_id, record_index, record_data, msisdn, is_valid, created_at)
                    VALUES
                    (:dataset_id, :record_index, :record_data::jsonb, :msisdn, TRUE, NOW())
                """), {
                    'dataset_id': dataset_id,
                    'record_index': idx,
                    'record_data': json.dumps(record_data, default=str),
                    'msisdn': record_data.get('msisdn', '')
                })

                valid_records += 1

            # Update dataset
            db.execute(text("""
                UPDATE tbl_dcdl_datasets
                SET total_records = :total_records,
                    valid_records = :valid_records,
                    status = 'COMPLETED',
                    last_refreshed_at = NOW(),
                    cache_expiry = NOW() + INTERVAL '7 days'
                WHERE dataset_id = :dataset_id
            """), {
                'dataset_id': dataset_id,
                'total_records': valid_records,
                'valid_records': valid_records
            })

            db.commit()

            logger.info(f"Refreshed dataset {dataset_id}: {valid_records} records")
            return {
                'total_records': valid_records,
                'valid_records': valid_records,
                'status': 'COMPLETED'
            }

        except Exception as e:
            db.rollback()
            logger.error(f"Error refreshing dataset: {str(e)}")
            raise

    def get_cached_data(
        self,
        db: Session,
        dataset_id: int,
        msisdn: Optional[str] = None,
        offset: int = 0,
        limit: int = 100
    ) -> Tuple[List[Dict], int]:
        """
        Get cached dataset records

        Args:
            db: Database session
            dataset_id: Dataset ID
            msisdn: Optional MSISDN filter
            offset: Pagination offset
            limit: Results per page

        Returns:
            Tuple of (records, total_count)
        """
        try:
            # Build query
            where_clause = "dataset_id = :dataset_id AND is_valid = TRUE"
            params = {'dataset_id': dataset_id}

            if msisdn:
                where_clause += " AND msisdn = :msisdn"
                params['msisdn'] = msisdn

            # Get total count
            count_result = db.execute(text(f"""
                SELECT COUNT(*) FROM tbl_dcdl_data_cache
                WHERE {where_clause}
            """), params).fetchone()

            total = count_result[0]

            # Get records
            results = db.execute(text(f"""
                SELECT record_data, msisdn, created_at
                FROM tbl_dcdl_data_cache
                WHERE {where_clause}
                ORDER BY record_index
                LIMIT :limit OFFSET :offset
            """), {**params, 'limit': limit, 'offset': offset}).fetchall()

            records = []
            for row in results:
                record_data = json.loads(row[0]) if isinstance(row[0], str) else row[0]
                records.append({
                    'data': record_data,
                    'msisdn': row[1],
                    'created_at': row[2].isoformat() if row[2] else None
                })

            return records, total

        except Exception as e:
            logger.error(f"Error fetching cached data: {str(e)}")
            raise

    def create_parameter_mapping(
        self,
        db: Session,
        dataset_id: int,
        mapping_config: Dict[str, Any]
    ) -> int:
        """
        Create parameter mapping for dataset

        Args:
            db: Database session
            dataset_id: Dataset ID
            mapping_config: Mapping configuration

        Returns:
            Mapping ID
        """
        try:
            result = db.execute(text("""
                INSERT INTO tbl_dcdl_mapping
                (dataset_id, source_column, parameter_name, transform_function,
                 default_value, placeholder, is_required, validation_regex,
                 created_at)
                VALUES
                (:dataset_id, :source_column, :parameter_name, :transform_function,
                 :default_value, :placeholder, :is_required, :validation_regex, NOW())
                RETURNING mapping_id
            """), {
                'dataset_id': dataset_id,
                'source_column': mapping_config.get('source_column'),
                'parameter_name': mapping_config.get('parameter_name'),
                'transform_function': mapping_config.get('transform_function'),
                'default_value': mapping_config.get('default_value'),
                'placeholder': mapping_config.get('placeholder'),
                'is_required': mapping_config.get('is_required', False),
                'validation_regex': mapping_config.get('validation_regex')
            })

            mapping_id = result.fetchone()[0]
            db.commit()

            logger.info(f"Created parameter mapping {mapping_id}")
            return mapping_id

        except Exception as e:
            db.rollback()
            logger.error(f"Error creating mapping: {str(e)}")
            raise

    def apply_parameter_mapping(
        self,
        record_data: Dict[str, Any],
        mappings: List[Dict[str, Any]]
    ) -> Dict[str, Any]:
        """
        Apply parameter mappings to record data

        Args:
            record_data: Raw record data
            mappings: List of mapping configurations

        Returns:
            Transformed record data
        """
        result = {}

        for mapping in mappings:
            source_col = mapping.get('source_column')
            param_name = mapping.get('parameter_name')
            transform_func = mapping.get('transform_function')
            default_value = mapping.get('default_value')

            # Get source value
            value = record_data.get(source_col, default_value)

            if value is not None:
                # Apply transformation
                if transform_func == 'UPPER':
                    value = str(value).upper()
                elif transform_func == 'LOWER':
                    value = str(value).lower()
                elif transform_func == 'TRIM':
                    value = str(value).strip()
                elif transform_func == 'FORMAT_MSISDN':
                    # Remove non-digits
                    value = ''.join(filter(str.isdigit, str(value)))

                result[param_name] = value

        return result

    def _generate_dataset_code(self, db: Session) -> str:
        """Generate unique dataset code"""
        import random
        import string

        while True:
            code = 'DCDL_' + ''.join(random.choices(string.ascii_uppercase + string.digits, k=12))
            existing = db.execute(text("""
                SELECT 1 FROM tbl_dcdl_datasets WHERE dataset_code = :code
            """), {'code': code}).fetchone()

            if not existing:
                return code

    def delete_dataset(
        self,
        db: Session,
        dataset_id: int
    ) -> bool:
        """
        Delete dataset and cached data

        Args:
            db: Database session
            dataset_id: Dataset ID

        Returns:
            True if successful
        """
        try:
            # Delete cached data
            db.execute(text("""
                DELETE FROM tbl_dcdl_data_cache WHERE dataset_id = :dataset_id
            """), {'dataset_id': dataset_id})

            # Delete mappings
            db.execute(text("""
                DELETE FROM tbl_dcdl_mapping WHERE dataset_id = :dataset_id
            """), {'dataset_id': dataset_id})

            # Delete queries
            db.execute(text("""
                DELETE FROM tbl_dcdl_queries WHERE dataset_id = :dataset_id
            """), {'dataset_id': dataset_id})

            # Delete dataset
            db.execute(text("""
                DELETE FROM tbl_dcdl_datasets WHERE dataset_id = :dataset_id
            """), {'dataset_id': dataset_id})

            db.commit()

            logger.info(f"Deleted dataset {dataset_id}")
            return True

        except Exception as e:
            db.rollback()
            logger.error(f"Error deleting dataset: {str(e)}")
            raise
