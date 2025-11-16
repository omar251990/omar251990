#!/usr/bin/env python3
"""
Profile Management API Endpoints
Subscriber profiles with privacy-first design
"""

from fastapi import APIRouter, Depends, HTTPException, Query, UploadFile, File, Response
from sqlalchemy.orm import Session
from typing import Optional, List, Dict, Any
from datetime import datetime, date
from pydantic import BaseModel, Field

from src.core.database import get_db
from src.services.auth import get_current_user
from src.models.user import User
from src.services.profile_service import ProfileService, AttributeSchemaService
from src.services.profile_import_export import ProfileImportService, ProfileExportService

router = APIRouter(prefix="/profiles", tags=["profiles"])


# ========== Pydantic Models ==========

class ProfileCreate(BaseModel):
    """Profile creation request"""
    msisdn: str = Field(..., description="Subscriber MSISDN (will be hashed)")
    gender: Optional[str] = None
    age: Optional[int] = None
    date_of_birth: Optional[date] = None
    language: Optional[str] = None
    country_code: Optional[str] = None
    region: Optional[str] = None
    city: Optional[str] = None
    postal_code: Optional[str] = None
    device_type: Optional[str] = None
    device_model: Optional[str] = None
    os_version: Optional[str] = None
    plan_type: Optional[str] = None
    subscription_date: Optional[date] = None
    last_recharge_date: Optional[date] = None
    interests: Optional[List[str]] = None
    preferences: Optional[Dict[str, Any]] = None
    opt_in_marketing: Optional[bool] = False
    opt_in_sms: Optional[bool] = True
    custom_attributes: Optional[Dict[str, Any]] = None


class ProfileUpdate(BaseModel):
    """Profile update request"""
    gender: Optional[str] = None
    age: Optional[int] = None
    date_of_birth: Optional[date] = None
    language: Optional[str] = None
    region: Optional[str] = None
    city: Optional[str] = None
    device_type: Optional[str] = None
    plan_type: Optional[str] = None
    opt_in_marketing: Optional[bool] = None
    opt_in_sms: Optional[bool] = None
    custom_attributes: Optional[Dict[str, Any]] = None


class ProfileSearch(BaseModel):
    """Profile search filters"""
    gender: Optional[str] = None
    age_min: Optional[int] = None
    age_max: Optional[int] = None
    region: Optional[str] = None
    city: Optional[str] = None
    device_type: Optional[str] = None
    plan_type: Optional[str] = None
    status: Optional[str] = None
    opt_in_marketing: Optional[bool] = None
    last_activity_after: Optional[date] = None
    last_activity_before: Optional[date] = None
    custom_attributes: Optional[Dict[str, Any]] = None


class ImportJobCreate(BaseModel):
    """Import job creation request"""
    file_type: str = Field(..., description="File type: CSV, EXCEL, JSON")
    column_mapping: Dict[str, str] = Field(..., description="Column to field mapping")
    update_existing: Optional[bool] = True
    skip_duplicates: Optional[bool] = False
    hash_msisdn: Optional[bool] = True


class AttributeSchemaCreate(BaseModel):
    """Attribute schema creation request"""
    attribute_name: str
    attribute_code: str
    display_name: str
    description: Optional[str] = None
    data_type: str = Field(..., description="STRING, INTEGER, DECIMAL, BOOLEAN, ENUM, JSON, DATE, DATETIME")
    allowed_values: Optional[List[Any]] = None
    validation_regex: Optional[str] = None
    min_value: Optional[float] = None
    max_value: Optional[float] = None
    is_required: Optional[bool] = False
    is_searchable: Optional[bool] = True
    is_visible_to_cp: Optional[bool] = True
    is_encrypted: Optional[bool] = False
    privacy_level: Optional[str] = 'PUBLIC'
    display_order: Optional[int] = 100
    category: Optional[str] = None


# ========== Service Instances ==========

profile_service = ProfileService()
attribute_service = AttributeSchemaService()
import_service = ProfileImportService()
export_service = ProfileExportService()


# ========== Profile CRUD Endpoints ==========

@router.post("/")
async def create_profile(
    profile_data: ProfileCreate,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """
    Create a new subscriber profile

    - **msisdn**: Subscriber MSISDN (will be hashed for privacy)
    - **custom_attributes**: Any additional custom fields
    """
    try:
        # Get customer_id from current user
        customer_id = current_user.customer_id

        # Convert Pydantic model to dict
        data = profile_data.dict(exclude_unset=True)
        msisdn = data.pop('msisdn')

        # Create profile
        profile = profile_service.create_profile(
            db=db,
            msisdn=msisdn,
            customer_id=customer_id,
            profile_data=data,
            user_id=current_user.id
        )

        return {
            "status": "success",
            "message": "Profile created successfully",
            "data": {
                "profile_id": profile.profile_id,
                "msisdn_hash": profile.msisdn_hash,
                "created_at": profile.created_at.isoformat()
            }
        }

    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get("/{profile_id}")
async def get_profile(
    profile_id: int,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """Get profile by ID"""
    try:
        profile = profile_service.get_profile(db=db, profile_id=profile_id)

        if not profile:
            raise HTTPException(status_code=404, detail="Profile not found")

        # Check customer ownership
        if profile.customer_id != current_user.customer_id:
            raise HTTPException(status_code=403, detail="Access denied")

        return {
            "status": "success",
            "data": {
                "profile_id": profile.profile_id,
                "msisdn_hash": profile.msisdn_hash,
                "gender": profile.gender,
                "age": profile.age,
                "language": profile.language,
                "region": profile.region,
                "city": profile.city,
                "device_type": profile.device_type,
                "plan_type": profile.plan_type,
                "status": profile.status,
                "opt_in_marketing": profile.opt_in_marketing,
                "opt_in_sms": profile.opt_in_sms,
                "custom_attributes": profile.custom_attributes,
                "last_activity_date": profile.last_activity_date.isoformat() if profile.last_activity_date else None,
                "created_at": profile.created_at.isoformat(),
                "updated_at": profile.updated_at.isoformat()
            }
        }

    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.put("/{profile_id}")
async def update_profile(
    profile_id: int,
    profile_data: ProfileUpdate,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """Update existing profile"""
    try:
        # Verify profile exists and belongs to customer
        existing = profile_service.get_profile(db=db, profile_id=profile_id)

        if not existing:
            raise HTTPException(status_code=404, detail="Profile not found")

        if existing.customer_id != current_user.customer_id:
            raise HTTPException(status_code=403, detail="Access denied")

        # Update profile
        data = profile_data.dict(exclude_unset=True)
        profile = profile_service.update_profile(
            db=db,
            profile_id=profile_id,
            profile_data=data,
            user_id=current_user.id
        )

        return {
            "status": "success",
            "message": "Profile updated successfully",
            "data": {
                "profile_id": profile.profile_id,
                "updated_at": profile.updated_at.isoformat()
            }
        }

    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.delete("/{profile_id}")
async def delete_profile(
    profile_id: int,
    soft_delete: bool = Query(True, description="Soft delete (mark as deleted) or hard delete"),
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """Delete profile"""
    try:
        # Verify profile exists and belongs to customer
        existing = profile_service.get_profile(db=db, profile_id=profile_id)

        if not existing:
            raise HTTPException(status_code=404, detail="Profile not found")

        if existing.customer_id != current_user.customer_id:
            raise HTTPException(status_code=403, detail="Access denied")

        # Delete profile
        success = profile_service.delete_profile(
            db=db,
            profile_id=profile_id,
            soft_delete=soft_delete
        )

        return {
            "status": "success",
            "message": f"Profile {'soft' if soft_delete else 'hard'} deleted successfully"
        }

    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


# ========== Profile Search Endpoints ==========

@router.post("/search")
async def search_profiles(
    filters: ProfileSearch,
    offset: int = Query(0, ge=0),
    limit: int = Query(100, ge=1, le=1000),
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """
    Search profiles with filters

    Supports filtering by demographics, location, device, plan type, and custom attributes.
    """
    try:
        customer_id = current_user.customer_id

        # Convert filters to dict
        filter_dict = filters.dict(exclude_unset=True)

        # Search profiles
        profiles, total = profile_service.search_profiles(
            db=db,
            customer_id=customer_id,
            filters=filter_dict,
            user_id=current_user.id,
            offset=offset,
            limit=limit
        )

        # Format results
        results = []
        for profile in profiles:
            results.append({
                "profile_id": profile.profile_id,
                "msisdn_hash": profile.msisdn_hash,
                "gender": profile.gender,
                "age": profile.age,
                "region": profile.region,
                "city": profile.city,
                "device_type": profile.device_type,
                "plan_type": profile.plan_type,
                "status": profile.status,
                "opt_in_marketing": profile.opt_in_marketing
            })

        return {
            "status": "success",
            "data": {
                "profiles": results,
                "total": total,
                "offset": offset,
                "limit": limit
            }
        }

    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


# ========== Bulk Operations Endpoints ==========

@router.post("/bulk/update")
async def bulk_update_profiles(
    profile_ids: List[int],
    update_data: ProfileUpdate,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """Bulk update multiple profiles"""
    try:
        data = update_data.dict(exclude_unset=True)

        count = profile_service.bulk_update_profiles(
            db=db,
            profile_ids=profile_ids,
            update_data=data,
            user_id=current_user.id
        )

        return {
            "status": "success",
            "message": f"Updated {count} profiles",
            "data": {
                "updated_count": count
            }
        }

    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


# ========== Statistics Endpoints ==========

@router.get("/statistics/current")
async def get_current_statistics(
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """Get current profile statistics"""
    try:
        customer_id = current_user.customer_id

        # Get or calculate statistics
        stats = profile_service.get_profile_statistics(db=db, customer_id=customer_id)

        if not stats:
            # Calculate fresh statistics
            stats = profile_service.calculate_profile_statistics(db=db, customer_id=customer_id)

        return {
            "status": "success",
            "data": {
                "total_profiles": stats.total_profiles,
                "active_profiles": stats.active_profiles,
                "inactive_profiles": stats.inactive_profiles,
                "demographics": {
                    "male_count": stats.male_count,
                    "female_count": stats.female_count,
                    "avg_age": float(stats.avg_age) if stats.avg_age else None
                },
                "device_distribution": {
                    "android_count": stats.android_count,
                    "ios_count": stats.ios_count,
                    "feature_phone_count": stats.feature_phone_count
                },
                "plan_distribution": {
                    "prepaid_count": stats.prepaid_count,
                    "postpaid_count": stats.postpaid_count
                },
                "opt_in_rates": {
                    "marketing": stats.opt_in_marketing_count,
                    "sms": stats.opt_in_sms_count
                },
                "period_date": stats.period_date.isoformat()
            }
        }

    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


# ========== Import/Export Endpoints ==========

@router.post("/import")
async def import_profiles(
    file: UploadFile = File(...),
    job_config: str = Query(..., description="JSON config for import job"),
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """
    Import profiles from file (CSV, Excel, JSON)

    - Upload file and provide column mapping
    - Returns import job ID for tracking
    """
    try:
        import json
        config = json.loads(job_config)

        customer_id = current_user.customer_id

        # Validate file type
        file_ext = file.filename.split('.')[-1].upper()
        if file_ext not in ['CSV', 'XLSX', 'JSON']:
            raise HTTPException(status_code=400, detail="Unsupported file type")

        # Create import job
        job = import_service.create_import_job(
            db=db,
            customer_id=customer_id,
            user_id=current_user.id,
            file_name=file.filename,
            file_type=file_ext,
            column_mapping=config.get('column_mapping', {}),
            options=config.get('options', {})
        )

        # Read file content
        file_content = await file.read()

        # Process import based on file type
        if file_ext == 'CSV':
            job = import_service.import_from_csv(db=db, job_id=job.job_id, file_content=file_content)
        elif file_ext in ['XLSX', 'XLS']:
            job = import_service.import_from_excel(db=db, job_id=job.job_id, file_content=file_content)
        elif file_ext == 'JSON':
            job = import_service.import_from_json(db=db, job_id=job.job_id, file_content=file_content)

        return {
            "status": "success",
            "message": "Import completed",
            "data": {
                "job_id": job.job_id,
                "job_code": job.job_code,
                "status": job.status,
                "total_rows": job.total_rows,
                "rows_imported": job.rows_imported,
                "rows_updated": job.rows_updated,
                "rows_failed": job.rows_failed,
                "duration_seconds": job.duration_seconds
            }
        }

    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get("/import/jobs/{job_id}")
async def get_import_job(
    job_id: int,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """Get import job status"""
    try:
        job = import_service.get_import_job(db=db, job_id=job_id)

        if not job:
            raise HTTPException(status_code=404, detail="Import job not found")

        if job.customer_id != current_user.customer_id:
            raise HTTPException(status_code=403, detail="Access denied")

        return {
            "status": "success",
            "data": {
                "job_id": job.job_id,
                "job_code": job.job_code,
                "status": job.status,
                "file_name": job.file_name,
                "total_rows": job.total_rows,
                "rows_processed": job.rows_processed,
                "rows_imported": job.rows_imported,
                "rows_updated": job.rows_updated,
                "rows_failed": job.rows_failed,
                "error_message": job.error_message,
                "error_rows": job.error_rows,
                "started_at": job.started_at.isoformat() if job.started_at else None,
                "completed_at": job.completed_at.isoformat() if job.completed_at else None,
                "duration_seconds": job.duration_seconds
            }
        }

    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


# ========== Attribute Schema Endpoints ==========

@router.post("/attributes/schema")
async def create_attribute_schema(
    attribute_data: AttributeSchemaCreate,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """Create a new custom attribute schema"""
    try:
        # Only admin users can create attribute schemas
        if not current_user.is_admin:
            raise HTTPException(status_code=403, detail="Admin access required")

        data = attribute_data.dict()
        attribute = attribute_service.create_attribute(
            db=db,
            attribute_data=data,
            user_id=str(current_user.id)
        )

        return {
            "status": "success",
            "message": "Attribute schema created successfully",
            "data": {
                "attribute_id": attribute.attribute_id,
                "attribute_code": attribute.attribute_code,
                "attribute_name": attribute.attribute_name
            }
        }

    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get("/attributes/schema")
async def get_attribute_schemas(
    active_only: bool = Query(True, description="Return only active attributes"),
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """Get all attribute schemas"""
    try:
        attributes = attribute_service.get_attributes(db=db, active_only=active_only)

        results = []
        for attr in attributes:
            results.append({
                "attribute_id": attr.attribute_id,
                "attribute_name": attr.attribute_name,
                "attribute_code": attr.attribute_code,
                "display_name": attr.display_name,
                "data_type": attr.data_type,
                "is_required": attr.is_required,
                "is_searchable": attr.is_searchable,
                "privacy_level": attr.privacy_level,
                "display_order": attr.display_order,
                "category": attr.category
            })

        return {
            "status": "success",
            "data": {
                "attributes": results,
                "total": len(results)
            }
        }

    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))
