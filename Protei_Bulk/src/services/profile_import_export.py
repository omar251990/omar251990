#!/usr/bin/env python3
"""
Profile Import/Export Service
Bulk profile operations with file support (CSV, Excel, JSON)
"""

import csv
import json
import logging
import pandas as pd
from typing import List, Dict, Any, Optional, Tuple
from datetime import datetime
from io import StringIO, BytesIO
from pathlib import Path

from sqlalchemy import and_
from sqlalchemy.orm import Session

from src.models.profiling import Profile, ProfileImportJob, ProfileGroup, ProfileGroupMember
from src.services.profile_service import ProfileService

logger = logging.getLogger(__name__)


class ProfileImportService:
    """Service for importing profiles from files"""

    def __init__(self):
        self.profile_service = ProfileService()

    def create_import_job(
        self,
        db: Session,
        customer_id: int,
        user_id: int,
        file_name: str,
        file_type: str,
        column_mapping: Dict[str, str],
        options: Optional[Dict[str, Any]] = None
    ) -> ProfileImportJob:
        """
        Create a new import job

        Args:
            db: Database session
            customer_id: Customer ID
            user_id: User creating the job
            file_name: Original file name
            file_type: File type (CSV, EXCEL, JSON)
            column_mapping: Mapping of file columns to profile fields
            options: Import options (update_existing, skip_duplicates, hash_msisdn)

        Returns:
            Created ProfileImportJob object
        """
        try:
            # Generate job code
            job_code = self._generate_job_code(db)

            # Default options
            if options is None:
                options = {}

            job = ProfileImportJob(
                job_code=job_code,
                customer_id=customer_id,
                user_id=user_id,
                file_name=file_name,
                file_type=file_type.upper(),
                column_mapping=column_mapping,
                update_existing=options.get('update_existing', True),
                skip_duplicates=options.get('skip_duplicates', False),
                hash_msisdn=options.get('hash_msisdn', True),
                created_by=str(user_id),
                status='PENDING'
            )

            db.add(job)
            db.commit()
            db.refresh(job)

            logger.info(f"Created import job {job.job_code}")
            return job

        except Exception as e:
            db.rollback()
            logger.error(f"Error creating import job: {str(e)}")
            raise

    def import_from_csv(
        self,
        db: Session,
        job_id: int,
        file_content: bytes
    ) -> ProfileImportJob:
        """
        Import profiles from CSV file

        Args:
            db: Database session
            job_id: Import job ID
            file_content: CSV file content as bytes

        Returns:
            Updated ProfileImportJob object
        """
        try:
            job = db.query(ProfileImportJob).filter(ProfileImportJob.job_id == job_id).first()

            if not job:
                raise ValueError(f"Import job {job_id} not found")

            # Update status
            job.status = 'PROCESSING'
            job.started_at = datetime.utcnow()
            db.commit()

            # Parse CSV
            csv_text = file_content.decode('utf-8')
            csv_reader = csv.DictReader(StringIO(csv_text))

            rows = list(csv_reader)
            job.total_rows = len(rows)

            # Process rows
            for idx, row in enumerate(rows):
                try:
                    # Map columns to profile fields
                    profile_data = self._map_row_to_profile(row, job.column_mapping)

                    # Extract MSISDN
                    msisdn = profile_data.pop('msisdn', None)
                    if not msisdn:
                        job.rows_failed += 1
                        self._add_error_row(job, idx, "Missing MSISDN", row)
                        continue

                    # Check if profile exists
                    existing = self.profile_service.get_profile(db, msisdn=msisdn)

                    if existing:
                        if job.update_existing:
                            # Update existing profile
                            self.profile_service.update_profile(
                                db, existing.profile_id, profile_data, job.user_id
                            )
                            job.rows_updated += 1
                        elif job.skip_duplicates:
                            # Skip duplicate
                            job.rows_processed += 1
                            continue
                        else:
                            job.rows_failed += 1
                            self._add_error_row(job, idx, "Duplicate MSISDN", row)
                            continue
                    else:
                        # Create new profile
                        self.profile_service.create_profile(
                            db, msisdn, job.customer_id, profile_data, job.user_id
                        )
                        job.rows_imported += 1

                    job.rows_processed += 1

                except Exception as row_error:
                    job.rows_failed += 1
                    self._add_error_row(job, idx, str(row_error), row)
                    logger.warning(f"Error processing row {idx}: {str(row_error)}")

            # Update job status
            job.status = 'COMPLETED' if job.rows_failed == 0 else 'COMPLETED'
            job.completed_at = datetime.utcnow()
            job.duration_seconds = (job.completed_at - job.started_at).total_seconds()

            db.commit()
            db.refresh(job)

            logger.info(f"Import job {job.job_code} completed: {job.rows_imported} imported, {job.rows_updated} updated, {job.rows_failed} failed")
            return job

        except Exception as e:
            if job:
                job.status = 'FAILED'
                job.error_message = str(e)
                job.completed_at = datetime.utcnow()
                db.commit()

            logger.error(f"Error importing CSV: {str(e)}")
            raise

    def import_from_excel(
        self,
        db: Session,
        job_id: int,
        file_content: bytes
    ) -> ProfileImportJob:
        """
        Import profiles from Excel file

        Args:
            db: Database session
            job_id: Import job ID
            file_content: Excel file content as bytes

        Returns:
            Updated ProfileImportJob object
        """
        try:
            job = db.query(ProfileImportJob).filter(ProfileImportJob.job_id == job_id).first()

            if not job:
                raise ValueError(f"Import job {job_id} not found")

            # Update status
            job.status = 'PROCESSING'
            job.started_at = datetime.utcnow()
            db.commit()

            # Parse Excel using pandas
            df = pd.read_excel(BytesIO(file_content))

            job.total_rows = len(df)

            # Process rows
            for idx, row in df.iterrows():
                try:
                    # Convert row to dict
                    row_dict = row.to_dict()

                    # Map columns to profile fields
                    profile_data = self._map_row_to_profile(row_dict, job.column_mapping)

                    # Extract MSISDN
                    msisdn = profile_data.pop('msisdn', None)
                    if not msisdn:
                        job.rows_failed += 1
                        self._add_error_row(job, idx, "Missing MSISDN", row_dict)
                        continue

                    # Convert NaN to None
                    profile_data = {k: (None if pd.isna(v) else v) for k, v in profile_data.items()}

                    # Check if profile exists
                    existing = self.profile_service.get_profile(db, msisdn=str(msisdn))

                    if existing:
                        if job.update_existing:
                            self.profile_service.update_profile(
                                db, existing.profile_id, profile_data, job.user_id
                            )
                            job.rows_updated += 1
                        elif job.skip_duplicates:
                            job.rows_processed += 1
                            continue
                        else:
                            job.rows_failed += 1
                            self._add_error_row(job, idx, "Duplicate MSISDN", row_dict)
                            continue
                    else:
                        self.profile_service.create_profile(
                            db, str(msisdn), job.customer_id, profile_data, job.user_id
                        )
                        job.rows_imported += 1

                    job.rows_processed += 1

                except Exception as row_error:
                    job.rows_failed += 1
                    self._add_error_row(job, idx, str(row_error), row_dict)
                    logger.warning(f"Error processing row {idx}: {str(row_error)}")

            # Update job status
            job.status = 'COMPLETED'
            job.completed_at = datetime.utcnow()
            job.duration_seconds = (job.completed_at - job.started_at).total_seconds()

            db.commit()
            db.refresh(job)

            logger.info(f"Excel import job {job.job_code} completed")
            return job

        except Exception as e:
            if job:
                job.status = 'FAILED'
                job.error_message = str(e)
                job.completed_at = datetime.utcnow()
                db.commit()

            logger.error(f"Error importing Excel: {str(e)}")
            raise

    def import_from_json(
        self,
        db: Session,
        job_id: int,
        file_content: bytes
    ) -> ProfileImportJob:
        """
        Import profiles from JSON file

        Args:
            db: Database session
            job_id: Import job ID
            file_content: JSON file content as bytes

        Returns:
            Updated ProfileImportJob object
        """
        try:
            job = db.query(ProfileImportJob).filter(ProfileImportJob.job_id == job_id).first()

            if not job:
                raise ValueError(f"Import job {job_id} not found")

            # Update status
            job.status = 'PROCESSING'
            job.started_at = datetime.utcnow()
            db.commit()

            # Parse JSON
            json_data = json.loads(file_content.decode('utf-8'))

            # Handle different JSON structures
            if isinstance(json_data, list):
                rows = json_data
            elif isinstance(json_data, dict) and 'profiles' in json_data:
                rows = json_data['profiles']
            else:
                raise ValueError("Invalid JSON structure")

            job.total_rows = len(rows)

            # Process rows
            for idx, row in enumerate(rows):
                try:
                    # Map columns to profile fields
                    profile_data = self._map_row_to_profile(row, job.column_mapping)

                    # Extract MSISDN
                    msisdn = profile_data.pop('msisdn', None)
                    if not msisdn:
                        job.rows_failed += 1
                        self._add_error_row(job, idx, "Missing MSISDN", row)
                        continue

                    # Check if profile exists
                    existing = self.profile_service.get_profile(db, msisdn=msisdn)

                    if existing:
                        if job.update_existing:
                            self.profile_service.update_profile(
                                db, existing.profile_id, profile_data, job.user_id
                            )
                            job.rows_updated += 1
                        elif job.skip_duplicates:
                            job.rows_processed += 1
                            continue
                        else:
                            job.rows_failed += 1
                            self._add_error_row(job, idx, "Duplicate MSISDN", row)
                            continue
                    else:
                        self.profile_service.create_profile(
                            db, msisdn, job.customer_id, profile_data, job.user_id
                        )
                        job.rows_imported += 1

                    job.rows_processed += 1

                except Exception as row_error:
                    job.rows_failed += 1
                    self._add_error_row(job, idx, str(row_error), row)
                    logger.warning(f"Error processing row {idx}: {str(row_error)}")

            # Update job status
            job.status = 'COMPLETED'
            job.completed_at = datetime.utcnow()
            job.duration_seconds = (job.completed_at - job.started_at).total_seconds()

            db.commit()
            db.refresh(job)

            logger.info(f"JSON import job {job.job_code} completed")
            return job

        except Exception as e:
            if job:
                job.status = 'FAILED'
                job.error_message = str(e)
                job.completed_at = datetime.utcnow()
                db.commit()

            logger.error(f"Error importing JSON: {str(e)}")
            raise

    def get_import_job(
        self,
        db: Session,
        job_id: Optional[int] = None,
        job_code: Optional[str] = None
    ) -> Optional[ProfileImportJob]:
        """Get import job by ID or code"""
        try:
            if job_id:
                return db.query(ProfileImportJob).filter(ProfileImportJob.job_id == job_id).first()
            elif job_code:
                return db.query(ProfileImportJob).filter(ProfileImportJob.job_code == job_code).first()
            return None

        except Exception as e:
            logger.error(f"Error fetching import job: {str(e)}")
            raise

    def _map_row_to_profile(
        self,
        row: Dict[str, Any],
        column_mapping: Dict[str, str]
    ) -> Dict[str, Any]:
        """
        Map file row to profile fields using column mapping

        Args:
            row: Source row data
            column_mapping: Mapping of source columns to profile fields

        Returns:
            Mapped profile data
        """
        profile_data = {}

        for source_col, target_field in column_mapping.items():
            if source_col in row:
                value = row[source_col]

                # Type conversions
                if target_field == 'age' and value:
                    try:
                        value = int(value)
                    except (ValueError, TypeError):
                        value = None

                if value is not None:
                    profile_data[target_field] = value

        return profile_data

    def _add_error_row(
        self,
        job: ProfileImportJob,
        row_index: int,
        error_message: str,
        row_data: Dict[str, Any]
    ) -> None:
        """Add error row to job"""
        if job.error_rows is None:
            job.error_rows = []

        # Limit error rows to prevent excessive data
        if len(job.error_rows) < 100:
            job.error_rows.append({
                'index': row_index,
                'error': error_message,
                'data': row_data
            })

    def _generate_job_code(self, db: Session) -> str:
        """Generate unique job code"""
        import random
        import string

        while True:
            code = 'IMP_' + ''.join(random.choices(string.ascii_uppercase + string.digits, k=12))
            existing = db.query(ProfileImportJob).filter(ProfileImportJob.job_code == code).first()
            if not existing:
                return code


class ProfileExportService:
    """Service for exporting profiles and segments"""

    def export_segment_to_csv(
        self,
        db: Session,
        group_id: int,
        include_fields: Optional[List[str]] = None
    ) -> str:
        """
        Export segment members to CSV

        Args:
            db: Database session
            group_id: Segment ID
            include_fields: List of fields to include (None = all)

        Returns:
            CSV content as string
        """
        try:
            # Get segment
            segment = db.query(ProfileGroup).filter(ProfileGroup.group_id == group_id).first()

            if not segment:
                raise ValueError(f"Segment {group_id} not found")

            # Get members
            profiles = db.query(Profile).join(
                ProfileGroupMember,
                Profile.profile_id == ProfileGroupMember.profile_id
            ).filter(
                ProfileGroupMember.group_id == group_id
            ).all()

            # Determine fields to export
            if include_fields is None:
                include_fields = [
                    'profile_id', 'gender', 'age', 'language',
                    'country_code', 'region', 'city',
                    'device_type', 'plan_type', 'status',
                    'opt_in_marketing', 'opt_in_sms'
                ]

            # Create CSV
            output = StringIO()
            writer = csv.DictWriter(output, fieldnames=include_fields)
            writer.writeheader()

            for profile in profiles:
                row = {}
                for field in include_fields:
                    value = getattr(profile, field, None)
                    row[field] = value if value is not None else ''

                writer.writerow(row)

            csv_content = output.getvalue()
            output.close()

            logger.info(f"Exported {len(profiles)} profiles from segment {group_id} to CSV")
            return csv_content

        except Exception as e:
            logger.error(f"Error exporting to CSV: {str(e)}")
            raise

    def export_segment_to_excel(
        self,
        db: Session,
        group_id: int,
        include_fields: Optional[List[str]] = None
    ) -> bytes:
        """
        Export segment members to Excel

        Args:
            db: Database session
            group_id: Segment ID
            include_fields: List of fields to include

        Returns:
            Excel content as bytes
        """
        try:
            # Get segment
            segment = db.query(ProfileGroup).filter(ProfileGroup.group_id == group_id).first()

            if not segment:
                raise ValueError(f"Segment {group_id} not found")

            # Get members
            profiles = db.query(Profile).join(
                ProfileGroupMember,
                Profile.profile_id == ProfileGroupMember.profile_id
            ).filter(
                ProfileGroupMember.group_id == group_id
            ).all()

            # Determine fields
            if include_fields is None:
                include_fields = [
                    'profile_id', 'gender', 'age', 'language',
                    'country_code', 'region', 'city',
                    'device_type', 'plan_type', 'status'
                ]

            # Create DataFrame
            data = []
            for profile in profiles:
                row = {}
                for field in include_fields:
                    row[field] = getattr(profile, field, None)
                data.append(row)

            df = pd.DataFrame(data)

            # Write to Excel
            output = BytesIO()
            with pd.ExcelWriter(output, engine='openpyxl') as writer:
                df.to_excel(writer, sheet_name='Profiles', index=False)

            excel_content = output.getvalue()
            output.close()

            logger.info(f"Exported {len(profiles)} profiles from segment {group_id} to Excel")
            return excel_content

        except Exception as e:
            logger.error(f"Error exporting to Excel: {str(e)}")
            raise

    def export_segment_to_json(
        self,
        db: Session,
        group_id: int,
        include_fields: Optional[List[str]] = None
    ) -> str:
        """
        Export segment members to JSON

        Args:
            db: Database session
            group_id: Segment ID
            include_fields: List of fields to include

        Returns:
            JSON content as string
        """
        try:
            # Get segment
            segment = db.query(ProfileGroup).filter(ProfileGroup.group_id == group_id).first()

            if not segment:
                raise ValueError(f"Segment {group_id} not found")

            # Get members
            profiles = db.query(Profile).join(
                ProfileGroupMember,
                Profile.profile_id == ProfileGroupMember.profile_id
            ).filter(
                ProfileGroupMember.group_id == group_id
            ).all()

            # Determine fields
            if include_fields is None:
                include_fields = [
                    'profile_id', 'gender', 'age', 'language',
                    'country_code', 'region', 'city',
                    'device_type', 'plan_type', 'status',
                    'custom_attributes'
                ]

            # Create JSON structure
            data = {
                'segment': {
                    'group_id': segment.group_id,
                    'group_name': segment.group_name,
                    'record_count': segment.record_count
                },
                'exported_at': datetime.utcnow().isoformat(),
                'profiles': []
            }

            for profile in profiles:
                profile_dict = {}
                for field in include_fields:
                    value = getattr(profile, field, None)

                    # Handle datetime serialization
                    if isinstance(value, datetime):
                        value = value.isoformat()

                    profile_dict[field] = value

                data['profiles'].append(profile_dict)

            json_content = json.dumps(data, indent=2, default=str)

            logger.info(f"Exported {len(profiles)} profiles from segment {group_id} to JSON")
            return json_content

        except Exception as e:
            logger.error(f"Error exporting to JSON: {str(e)}")
            raise
