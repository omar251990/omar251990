#!/usr/bin/env python3
"""
Segmentation API Endpoints
Dynamic audience segmentation and query builder
"""

from fastapi import APIRouter, Depends, HTTPException, Query, Response
from sqlalchemy.orm import Session
from typing import Optional, List, Dict, Any
from datetime import datetime
from pydantic import BaseModel, Field

from src.core.database import get_db
from src.services.auth import get_current_user
from src.models.user import User
from src.services.segmentation_service import SegmentationService, QueryBuilderService
from src.services.profile_import_export import ProfileExportService

router = APIRouter(prefix="/segments", tags=["segmentation"])


# ========== Pydantic Models ==========

class SegmentCreate(BaseModel):
    """Segment creation request"""
    group_name: str = Field(..., description="Segment name")
    description: Optional[str] = None
    filter_query: Dict[str, Any] = Field(..., description="Filter criteria")
    is_dynamic: Optional[bool] = True
    refresh_frequency: Optional[str] = 'DAILY'
    visibility: Optional[str] = 'PRIVATE'
    tags: Optional[List[str]] = None


class SegmentUpdate(BaseModel):
    """Segment update request"""
    group_name: Optional[str] = None
    description: Optional[str] = None
    filter_query: Optional[Dict[str, Any]] = None
    is_dynamic: Optional[bool] = None
    refresh_frequency: Optional[str] = None
    visibility: Optional[str] = None
    tags: Optional[List[str]] = None


class QueryValidation(BaseModel):
    """Query validation request"""
    query: Dict[str, Any] = Field(..., description="Query to validate")


class ExportRequest(BaseModel):
    """Export request"""
    format: str = Field(..., description="Export format: CSV, EXCEL, JSON")
    include_fields: Optional[List[str]] = None


# ========== Service Instances ==========

segmentation_service = SegmentationService()
query_builder_service = QueryBuilderService()
export_service = ProfileExportService()


# ========== Segment CRUD Endpoints ==========

@router.post("/")
async def create_segment(
    segment_data: SegmentCreate,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """
    Create a new segment

    - **group_name**: Name of the segment
    - **filter_query**: Filter criteria in query builder format
    - **is_dynamic**: If true, segment refreshes automatically
    - **refresh_frequency**: How often to refresh (REALTIME, HOURLY, DAILY, WEEKLY, MONTHLY, MANUAL)
    """
    try:
        customer_id = current_user.customer_id

        # Validate query structure
        is_valid, error = query_builder_service.validate_query(segment_data.filter_query)
        if not is_valid:
            raise HTTPException(status_code=400, detail=f"Invalid query: {error}")

        # Create segment
        data = segment_data.dict()
        segment = segmentation_service.create_segment(
            db=db,
            customer_id=customer_id,
            user_id=current_user.id,
            group_data=data
        )

        return {
            "status": "success",
            "message": "Segment created successfully",
            "data": {
                "group_id": segment.group_id,
                "group_code": segment.group_code,
                "group_name": segment.group_name,
                "record_count": segment.record_count,
                "created_at": segment.created_at.isoformat()
            }
        }

    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get("/{group_id}")
async def get_segment(
    group_id: int,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """Get segment details"""
    try:
        segment = segmentation_service.get_segment(db=db, group_id=group_id)

        if not segment:
            raise HTTPException(status_code=404, detail="Segment not found")

        # Check access
        if segment.customer_id != current_user.customer_id:
            # Check if shared
            if segment.visibility == 'PRIVATE':
                raise HTTPException(status_code=403, detail="Access denied")

        return {
            "status": "success",
            "data": {
                "group_id": segment.group_id,
                "group_code": segment.group_code,
                "group_name": segment.group_name,
                "description": segment.description,
                "filter_query": segment.filter_query,
                "filter_sql": segment.filter_sql,
                "record_count": segment.record_count,
                "is_dynamic": segment.is_dynamic,
                "refresh_frequency": segment.refresh_frequency,
                "last_refreshed": segment.last_refreshed.isoformat() if segment.last_refreshed else None,
                "next_refresh": segment.next_refresh.isoformat() if segment.next_refresh else None,
                "total_campaigns_sent": segment.total_campaigns_sent,
                "total_messages_sent": segment.total_messages_sent,
                "last_used_at": segment.last_used_at.isoformat() if segment.last_used_at else None,
                "visibility": segment.visibility,
                "is_active": segment.is_active,
                "tags": segment.tags,
                "created_at": segment.created_at.isoformat(),
                "updated_at": segment.updated_at.isoformat()
            }
        }

    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get("/")
async def list_segments(
    offset: int = Query(0, ge=0),
    limit: int = Query(100, ge=1, le=1000),
    is_active: Optional[bool] = Query(None, description="Filter by active status"),
    visibility: Optional[str] = Query(None, description="Filter by visibility"),
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """List all segments for the customer"""
    try:
        from src.models.profiling import ProfileGroup

        customer_id = current_user.customer_id

        # Build query
        query = db.query(ProfileGroup).filter(ProfileGroup.customer_id == customer_id)

        if is_active is not None:
            query = query.filter(ProfileGroup.is_active == is_active)

        if visibility:
            query = query.filter(ProfileGroup.visibility == visibility)

        # Get total count
        total = query.count()

        # Get segments
        segments = query.order_by(ProfileGroup.created_at.desc()).offset(offset).limit(limit).all()

        # Format results
        results = []
        for segment in segments:
            results.append({
                "group_id": segment.group_id,
                "group_code": segment.group_code,
                "group_name": segment.group_name,
                "description": segment.description,
                "record_count": segment.record_count,
                "is_dynamic": segment.is_dynamic,
                "refresh_frequency": segment.refresh_frequency,
                "last_refreshed": segment.last_refreshed.isoformat() if segment.last_refreshed else None,
                "visibility": segment.visibility,
                "is_active": segment.is_active,
                "created_at": segment.created_at.isoformat()
            })

        return {
            "status": "success",
            "data": {
                "segments": results,
                "total": total,
                "offset": offset,
                "limit": limit
            }
        }

    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.put("/{group_id}")
async def update_segment(
    group_id: int,
    segment_data: SegmentUpdate,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """Update segment"""
    try:
        # Verify segment exists and belongs to customer
        existing = segmentation_service.get_segment(db=db, group_id=group_id)

        if not existing:
            raise HTTPException(status_code=404, detail="Segment not found")

        if existing.customer_id != current_user.customer_id:
            raise HTTPException(status_code=403, detail="Access denied")

        # Validate query if provided
        if segment_data.filter_query:
            is_valid, error = query_builder_service.validate_query(segment_data.filter_query)
            if not is_valid:
                raise HTTPException(status_code=400, detail=f"Invalid query: {error}")

        # Update segment
        data = segment_data.dict(exclude_unset=True)
        segment = segmentation_service.update_segment(
            db=db,
            group_id=group_id,
            group_data=data,
            user_id=current_user.id
        )

        return {
            "status": "success",
            "message": "Segment updated successfully",
            "data": {
                "group_id": segment.group_id,
                "record_count": segment.record_count,
                "updated_at": segment.updated_at.isoformat()
            }
        }

    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.delete("/{group_id}")
async def delete_segment(
    group_id: int,
    soft_delete: bool = Query(True, description="Soft delete or hard delete"),
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """Delete segment"""
    try:
        # Verify segment exists and belongs to customer
        existing = segmentation_service.get_segment(db=db, group_id=group_id)

        if not existing:
            raise HTTPException(status_code=404, detail="Segment not found")

        if existing.customer_id != current_user.customer_id:
            raise HTTPException(status_code=403, detail="Access denied")

        # Delete segment
        success = segmentation_service.delete_segment(
            db=db,
            group_id=group_id,
            soft_delete=soft_delete
        )

        return {
            "status": "success",
            "message": f"Segment {'soft' if soft_delete else 'hard'} deleted successfully"
        }

    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


# ========== Segment Operations Endpoints ==========

@router.post("/{group_id}/refresh")
async def refresh_segment(
    group_id: int,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """
    Refresh segment membership

    Recalculates which profiles match the segment criteria
    """
    try:
        # Verify segment exists and belongs to customer
        existing = segmentation_service.get_segment(db=db, group_id=group_id)

        if not existing:
            raise HTTPException(status_code=404, detail="Segment not found")

        if existing.customer_id != current_user.customer_id:
            raise HTTPException(status_code=403, detail="Access denied")

        # Refresh segment
        total_members, new_members = segmentation_service.refresh_segment(
            db=db,
            group_id=group_id,
            user_id=current_user.id
        )

        return {
            "status": "success",
            "message": "Segment refreshed successfully",
            "data": {
                "total_members": total_members,
                "new_members": new_members,
                "refreshed_at": datetime.utcnow().isoformat()
            }
        }

    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get("/{group_id}/members")
async def get_segment_members(
    group_id: int,
    offset: int = Query(0, ge=0),
    limit: int = Query(100, ge=1, le=1000),
    include_profiles: bool = Query(False, description="Include full profile data"),
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """Get segment members"""
    try:
        # Verify segment exists and access
        segment = segmentation_service.get_segment(db=db, group_id=group_id)

        if not segment:
            raise HTTPException(status_code=404, detail="Segment not found")

        if segment.customer_id != current_user.customer_id:
            if segment.visibility == 'PRIVATE':
                raise HTTPException(status_code=403, detail="Access denied")

        # Get members
        members, total = segmentation_service.get_segment_members(
            db=db,
            group_id=group_id,
            offset=offset,
            limit=limit,
            include_profiles=include_profiles
        )

        # Format results
        if include_profiles:
            results = []
            for profile in members:
                results.append({
                    "profile_id": profile.profile_id,
                    "msisdn_hash": profile.msisdn_hash,
                    "gender": profile.gender,
                    "age": profile.age,
                    "region": profile.region,
                    "city": profile.city,
                    "device_type": profile.device_type,
                    "plan_type": profile.plan_type,
                    "status": profile.status
                })
        else:
            results = [
                {
                    "group_id": m.group_id,
                    "profile_id": m.profile_id,
                    "added_at": m.added_at.isoformat()
                }
                for m in members
            ]

        return {
            "status": "success",
            "data": {
                "members": results,
                "total": total,
                "offset": offset,
                "limit": limit
            }
        }

    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.post("/{group_id}/members/add")
async def add_members_to_segment(
    group_id: int,
    profile_ids: List[int],
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """Manually add profiles to a segment (non-dynamic segments only)"""
    try:
        # Verify segment exists and belongs to customer
        existing = segmentation_service.get_segment(db=db, group_id=group_id)

        if not existing:
            raise HTTPException(status_code=404, detail="Segment not found")

        if existing.customer_id != current_user.customer_id:
            raise HTTPException(status_code=403, detail="Access denied")

        # Add members
        count = segmentation_service.add_profiles_to_segment(
            db=db,
            group_id=group_id,
            profile_ids=profile_ids
        )

        return {
            "status": "success",
            "message": f"Added {count} profiles to segment",
            "data": {
                "added_count": count
            }
        }

    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.post("/{group_id}/members/remove")
async def remove_members_from_segment(
    group_id: int,
    profile_ids: List[int],
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """Manually remove profiles from a segment"""
    try:
        # Verify segment exists and belongs to customer
        existing = segmentation_service.get_segment(db=db, group_id=group_id)

        if not existing:
            raise HTTPException(status_code=404, detail="Segment not found")

        if existing.customer_id != current_user.customer_id:
            raise HTTPException(status_code=403, detail="Access denied")

        # Remove members
        count = segmentation_service.remove_profiles_from_segment(
            db=db,
            group_id=group_id,
            profile_ids=profile_ids
        )

        return {
            "status": "success",
            "message": f"Removed {count} profiles from segment",
            "data": {
                "removed_count": count
            }
        }

    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


# ========== Export Endpoints ==========

@router.post("/{group_id}/export")
async def export_segment(
    group_id: int,
    export_config: ExportRequest,
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """
    Export segment members

    - **format**: Export format (CSV, EXCEL, JSON)
    - **include_fields**: List of fields to include (optional)
    """
    try:
        # Verify segment exists and access
        segment = segmentation_service.get_segment(db=db, group_id=group_id)

        if not segment:
            raise HTTPException(status_code=404, detail="Segment not found")

        if segment.customer_id != current_user.customer_id:
            if segment.visibility == 'PRIVATE':
                raise HTTPException(status_code=403, detail="Access denied")

        # Export based on format
        format_type = export_config.format.upper()

        if format_type == 'CSV':
            content = export_service.export_segment_to_csv(
                db=db,
                group_id=group_id,
                include_fields=export_config.include_fields
            )
            media_type = 'text/csv'
            filename = f"segment_{segment.group_code}.csv"

        elif format_type == 'EXCEL':
            content = export_service.export_segment_to_excel(
                db=db,
                group_id=group_id,
                include_fields=export_config.include_fields
            )
            media_type = 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
            filename = f"segment_{segment.group_code}.xlsx"

        elif format_type == 'JSON':
            content = export_service.export_segment_to_json(
                db=db,
                group_id=group_id,
                include_fields=export_config.include_fields
            )
            media_type = 'application/json'
            filename = f"segment_{segment.group_code}.json"

        else:
            raise HTTPException(status_code=400, detail="Unsupported format")

        # Return file
        return Response(
            content=content,
            media_type=media_type,
            headers={
                'Content-Disposition': f'attachment; filename="{filename}"'
            }
        )

    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


# ========== Query Builder Endpoints ==========

@router.post("/query/validate")
async def validate_query(
    validation_request: QueryValidation,
    current_user: User = Depends(get_current_user)
):
    """
    Validate a query structure

    Checks if the query is properly formatted and can be executed
    """
    try:
        is_valid, error = query_builder_service.validate_query(validation_request.query)

        if is_valid:
            return {
                "status": "success",
                "data": {
                    "is_valid": True,
                    "message": "Query is valid"
                }
            }
        else:
            return {
                "status": "error",
                "data": {
                    "is_valid": False,
                    "error": error
                }
            }

    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.post("/query/preview")
async def preview_query(
    query: Dict[str, Any],
    limit: int = Query(10, ge=1, le=100),
    db: Session = Depends(get_db),
    current_user: User = Depends(get_current_user)
):
    """
    Preview query results without creating a segment

    Returns a sample of profiles matching the query
    """
    try:
        # Validate query
        is_valid, error = query_builder_service.validate_query(query)
        if not is_valid:
            raise HTTPException(status_code=400, detail=f"Invalid query: {error}")

        # Create temporary segment for preview
        temp_segment_data = {
            "group_name": f"TEMP_PREVIEW_{current_user.id}",
            "filter_query": query,
            "is_dynamic": False
        }

        segment = segmentation_service.create_segment(
            db=db,
            customer_id=current_user.customer_id,
            user_id=current_user.id,
            group_data=temp_segment_data
        )

        # Get sample members
        members, total = segmentation_service.get_segment_members(
            db=db,
            group_id=segment.group_id,
            offset=0,
            limit=limit,
            include_profiles=True
        )

        # Delete temporary segment
        segmentation_service.delete_segment(db=db, group_id=segment.group_id, soft_delete=False)

        # Format results
        results = []
        for profile in members:
            results.append({
                "profile_id": profile.profile_id,
                "gender": profile.gender,
                "age": profile.age,
                "region": profile.region,
                "city": profile.city,
                "device_type": profile.device_type,
                "plan_type": profile.plan_type
            })

        return {
            "status": "success",
            "data": {
                "preview": results,
                "total_matching": total,
                "showing": len(results)
            }
        }

    except HTTPException:
        raise
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))
