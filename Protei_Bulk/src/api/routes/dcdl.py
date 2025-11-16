#!/usr/bin/env python3
"""
DCDL (Dynamic Campaign Data Loader) API Endpoints
Manage datasets, queries, and parameter mappings
"""

from fastapi import APIRouter, Depends, HTTPException, Query, UploadFile, File, Response
from sqlalchemy.orm import Session
from typing import Optional, List, Dict, Any
from datetime import datetime
from pydantic import BaseModel, Field

from src.core.database import get_db
from src.services.auth import get_current_user
from src.models.user import User
from src.services.dcdl_service import DCDLService

router = APIRouter(prefix="/dcdl", tags=["dcdl"])


# ========== Pydantic Models ==========

class DatasetCreate(BaseModel):
    """Dataset creation request"""
    dataset_name: str = Field(..., description="Dataset name")
    description: Optional[str] = None
    source_type: str = Field(..., description="FILE_UPLOAD, DATABASE_QUERY, or HYBRID")
    file_type: Optional[str] = Field(None, description="CSV, EXCEL, JSON")


class DatabaseQueryCreate(BaseModel):
    """Database query dataset creation"""
    dataset_name: str
    description: Optional[str] = None
    query_type: str = Field(default="SQL", description="SQL or PROFILE_FILTER")
    query_text: str = Field(..., description="SQL query or filter JSON")
    parameters: Optional[Dict[str, Any]] = {}
    refresh_frequency: Optional[str] = Field("MANUAL", description="MANUAL, HOURLY, DAILY, WEEKLY")


class ParameterMappingCreate(BaseModel):
    """Parameter mapping creation"""
    source_column: str = Field(..., description="Source column name")
    parameter_name: str = Field(..., description="Target parameter name")
    transform_function: Optional[str] = Field(None, description="UPPER, LOWER, TRIM, FORMAT_MSISDN, etc.")
    default_value: Optional[str] = None
    placeholder: Optional[str] = Field(None, description="e.g., ${CUSTOMER_NAME}")
    is_required: Optional[bool] = False
    validation_regex: Optional[str] = None


# ========== Service Instance ==========

dcdl_service = DCDLService()


# ========== Dataset Management Endpoints ==========

@router.post("/datasets")
async def create_dataset(
    dataset_data: DatasetCreate,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """
    Create a new DCDL dataset

    - **dataset_name**: Name of the dataset
    - **source_type**: FILE_UPLOAD, DATABASE_QUERY, or HYBRID
    - **file_type**: CSV, EXCEL, or JSON (for file uploads)
    """
    try:
        customer_id = current_user.customer_id

        dataset = dcdl_service.create_dataset(
            db=db,
            customer_id=customer_id,
            user_id=current_user.id,
            dataset_config=dataset_data.dict()
        )

        return {
            "status": "success",
            "message": "Dataset created successfully",
            "data": dataset
        }

    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.post("/datasets/{dataset_id}/upload")
async def upload_dataset_file(
    dataset_id: int,
    file: UploadFile = File(...),
    column_mapping: str = Query(..., description="JSON string of column mapping"),
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """
    Upload data file to existing dataset

    - **file**: CSV, Excel, or JSON file
    - **column_mapping**: JSON mapping of file columns to parameters
    """
    try:
        import json
        mapping = json.loads(column_mapping)

        # Read file content
        file_content = await file.read()

        # Upload data
        result = dcdl_service.upload_csv_data(
            db=db,
            dataset_id=dataset_id,
            file_content=file_content,
            column_mapping=mapping
        )

        return {
            "status": "success",
            "message": "Data uploaded successfully",
            "data": result
        }

    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.post("/datasets/query")
async def create_query_dataset(
    query_data: DatabaseQueryCreate,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """
    Create dataset from database query

    - **query_text**: SQL query to execute
    - **refresh_frequency**: How often to refresh (MANUAL, HOURLY, DAILY, WEEKLY)
    """
    try:
        customer_id = current_user.customer_id

        dataset = dcdl_service.create_database_query_dataset(
            db=db,
            customer_id=customer_id,
            user_id=current_user.id,
            query_config=query_data.dict()
        )

        return {
            "status": "success",
            "message": "Query dataset created successfully",
            "data": dataset
        }

    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.post("/datasets/{dataset_id}/refresh")
async def refresh_dataset(
    dataset_id: int,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """
    Refresh dataset from source (re-execute query or reload file)
    """
    try:
        result = dcdl_service.refresh_query_dataset(
            db=db,
            dataset_id=dataset_id
        )

        return {
            "status": "success",
            "message": "Dataset refreshed successfully",
            "data": result
        }

    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get("/datasets/{dataset_id}/data")
async def get_dataset_data(
    dataset_id: int,
    msisdn: Optional[str] = Query(None, description="Filter by MSISDN"),
    offset: int = Query(0, ge=0),
    limit: int = Query(100, ge=1, le=1000),
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """
    Get cached dataset records

    - **msisdn**: Optional filter by MSISDN
    - **offset**: Pagination offset
    - **limit**: Results per page (max 1000)
    """
    try:
        records, total = dcdl_service.get_cached_data(
            db=db,
            dataset_id=dataset_id,
            msisdn=msisdn,
            offset=offset,
            limit=limit
        )

        return {
            "status": "success",
            "data": {
                "records": records,
                "total": total,
                "offset": offset,
                "limit": limit
            }
        }

    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.delete("/datasets/{dataset_id}")
async def delete_dataset(
    dataset_id: int,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """Delete dataset and all cached data"""
    try:
        success = dcdl_service.delete_dataset(
            db=db,
            dataset_id=dataset_id
        )

        return {
            "status": "success",
            "message": "Dataset deleted successfully"
        }

    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


# ========== Parameter Mapping Endpoints ==========

@router.post("/datasets/{dataset_id}/mappings")
async def create_parameter_mapping(
    dataset_id: int,
    mapping_data: ParameterMappingCreate,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """
    Create parameter mapping for dataset

    - **source_column**: Column name in source data
    - **parameter_name**: Target parameter name
    - **transform_function**: Optional transformation (UPPER, LOWER, TRIM, etc.)
    """
    try:
        mapping_id = dcdl_service.create_parameter_mapping(
            db=db,
            dataset_id=dataset_id,
            mapping_config=mapping_data.dict()
        )

        return {
            "status": "success",
            "message": "Mapping created successfully",
            "data": {
                "mapping_id": mapping_id
            }
        }

    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


# ========== Statistics Endpoints ==========

@router.get("/datasets/{dataset_id}/statistics")
async def get_dataset_statistics(
    dataset_id: int,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """Get dataset statistics and metadata"""
    try:
        from sqlalchemy import text

        result = db.execute(text("""
            SELECT
                dataset_code, dataset_name, source_type, file_type,
                total_records, valid_records, invalid_records,
                status, cache_expiry, created_at, last_refreshed_at
            FROM tbl_dcdl_datasets
            WHERE dataset_id = :dataset_id
        """), {'dataset_id': dataset_id}).fetchone()

        if not result:
            raise HTTPException(status_code=404, detail="Dataset not found")

        return {
            "status": "success",
            "data": {
                "dataset_code": result[0],
                "dataset_name": result[1],
                "source_type": result[2],
                "file_type": result[3],
                "total_records": result[4],
                "valid_records": result[5],
                "invalid_records": result[6],
                "status": result[7],
                "cache_expiry": result[8].isoformat() if result[8] else None,
                "created_at": result[9].isoformat() if result[9] else None,
                "last_refreshed_at": result[10].isoformat() if result[10] else None
            }
        }

    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get("/datasets")
async def list_datasets(
    source_type: Optional[str] = Query(None, description="Filter by source type"),
    status: Optional[str] = Query(None, description="Filter by status"),
    offset: int = Query(0, ge=0),
    limit: int = Query(100, ge=1, le=1000),
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """List all datasets for customer"""
    try:
        from sqlalchemy import text

        customer_id = current_user.customer_id

        where_clauses = ["customer_id = :customer_id"]
        params = {'customer_id': customer_id}

        if source_type:
            where_clauses.append("source_type = :source_type")
            params['source_type'] = source_type

        if status:
            where_clauses.append("status = :status")
            params['status'] = status

        where_sql = " AND ".join(where_clauses)

        # Get total count
        count_result = db.execute(text(f"""
            SELECT COUNT(*) FROM tbl_dcdl_datasets WHERE {where_sql}
        """), params).fetchone()

        total = count_result[0]

        # Get datasets
        results = db.execute(text(f"""
            SELECT
                dataset_id, dataset_code, dataset_name, source_type,
                file_type, total_records, valid_records, status,
                created_at, last_refreshed_at
            FROM tbl_dcdl_datasets
            WHERE {where_sql}
            ORDER BY created_at DESC
            LIMIT :limit OFFSET :offset
        """), {**params, 'limit': limit, 'offset': offset}).fetchall()

        datasets = []
        for row in results:
            datasets.append({
                "dataset_id": row[0],
                "dataset_code": row[1],
                "dataset_name": row[2],
                "source_type": row[3],
                "file_type": row[4],
                "total_records": row[5],
                "valid_records": row[6],
                "status": row[7],
                "created_at": row[8].isoformat() if row[8] else None,
                "last_refreshed_at": row[9].isoformat() if row[9] else None
            })

        return {
            "status": "success",
            "data": {
                "datasets": datasets,
                "total": total,
                "offset": offset,
                "limit": limit
            }
        }

    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))
