#!/usr/bin/env python3
"""
Bulk API Routes
HTTP API Gateway for message submission and DLR
"""

from typing import List, Optional, Dict, Any
from fastapi import APIRouter, Depends, HTTPException, Header, Request, status
from pydantic import BaseModel, Field, validator
from sqlalchemy.orm import Session
import secrets
from datetime import datetime

from src.core.database import get_db
from src.services.unified_auth_service import unified_auth_service
from src.models.user import User
from src.models.campaign import Campaign, Message
from src.models.quota import QuotaUsage


router = APIRouter(prefix="/api", tags=["Bulk API"])


# ============================================================================
# REQUEST/RESPONSE MODELS
# ============================================================================

class BulkMessageRecipient(BaseModel):
    """Individual recipient model"""
    msisdn: str = Field(..., description="Recipient phone number")
    message: Optional[str] = Field(None, description="Custom message for this recipient")
    variables: Optional[Dict[str, str]] = Field(default={}, description="Template variables")


class SendBulkRequest(BaseModel):
    """Bulk message submission request"""
    sender: str = Field(..., description="Sender ID", max_length=20)
    recipients: List[BulkMessageRecipient] = Field(..., description="List of recipients")
    message: str = Field(..., description="Default message content", max_length=1600)
    encoding: Optional[str] = Field("GSM7", description="Message encoding")
    priority: Optional[str] = Field("NORMAL", description="Message priority")
    dlr_url: Optional[str] = Field(None, description="DLR callback URL")
    campaign_name: Optional[str] = Field(None, description="Campaign name")
    schedule_time: Optional[datetime] = Field(None, description="Scheduled send time")

    @validator('recipients')
    def validate_recipients(cls, v):
        if len(v) == 0:
            raise ValueError('At least one recipient required')
        if len(v) > 100000:
            raise ValueError('Maximum 100,000 recipients per request')
        return v

    @validator('encoding')
    def validate_encoding(cls, v):
        if v not in ['GSM7', 'UCS2', 'ASCII']:
            raise ValueError('Invalid encoding. Must be GSM7, UCS2, or ASCII')
        return v


class SendBulkResponse(BaseModel):
    """Bulk message submission response"""
    status: str
    campaign_id: str
    total_recipients: int
    accepted: int
    rejected: int
    message: str
    errors: Optional[List[Dict[str, str]]] = []


class DLRQueryRequest(BaseModel):
    """DLR query request"""
    message_ids: Optional[List[str]] = Field(None, description="List of message IDs")
    campaign_id: Optional[str] = Field(None, description="Campaign ID")
    from_date: Optional[datetime] = Field(None, description="Start date")
    to_date: Optional[datetime] = Field(None, description="End date")
    limit: Optional[int] = Field(100, description="Maximum records to return")


class DLRRecord(BaseModel):
    """Delivery report record"""
    message_id: str
    msisdn: str
    status: str
    dlr_status: Optional[str]
    submitted_at: datetime
    delivered_at: Optional[datetime]
    error_code: Optional[str]


class DLRQueryResponse(BaseModel):
    """DLR query response"""
    status: str
    total: int
    records: List[DLRRecord]


class BalanceResponse(BaseModel):
    """Balance query response"""
    status: str
    balance: float
    credit_limit: float
    currency: str
    messages_sent_today: int
    messages_remaining_today: int
    daily_quota: int


# ============================================================================
# AUTHENTICATION DEPENDENCY
# ============================================================================

async def get_current_user_api_key(
    request: Request,
    authorization: Optional[str] = Header(None),
    db: Session = Depends(get_db)
) -> User:
    """
    Authenticate user using API key from Authorization header
    """
    if not authorization:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Authorization header required"
        )

    # Parse Authorization header
    parts = authorization.split()
    if len(parts) != 2 or parts[0].lower() != 'bearer':
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid authorization header format. Use: Bearer <api_key>"
        )

    api_key = parts[1]
    client_ip = request.client.host

    # Authenticate
    success, user, error = unified_auth_service.authenticate_api_key(
        db=db,
        api_key=api_key,
        ip_address=client_ip
    )

    if not success:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail=error.get('error', 'Authentication failed')
        )

    return user


async def get_current_user_basic_auth(
    request: Request,
    authorization: Optional[str] = Header(None),
    db: Session = Depends(get_db)
) -> User:
    """
    Authenticate user using Basic Auth
    """
    import base64

    if not authorization:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Authorization header required"
        )

    # Parse Authorization header
    parts = authorization.split()
    if len(parts) != 2 or parts[0].lower() != 'basic':
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid authorization header format. Use: Basic <credentials>"
        )

    try:
        credentials = base64.b64decode(parts[1]).decode('utf-8')
        username, password = credentials.split(':', 1)
    except Exception:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid Basic Auth credentials"
        )

    client_ip = request.client.host

    # Authenticate
    success, user, error = unified_auth_service.authenticate_basic_auth(
        db=db,
        username=username,
        password=password,
        ip_address=client_ip
    )

    if not success:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail=error.get('error', 'Authentication failed')
        )

    return user


# ============================================================================
# API ENDPOINTS
# ============================================================================

@router.post("/sendBulk", response_model=SendBulkResponse)
async def send_bulk_messages(
    request: SendBulkRequest,
    current_user: User = Depends(get_current_user_api_key),
    db: Session = Depends(get_db)
):
    """
    Submit bulk messages via HTTP API

    **Authentication**: API Key (Bearer token)

    **Request Body**:
    ```json
    {
      "sender": "MyCompany",
      "recipients": [
        {"msisdn": "962788123456", "message": "Hello Ahmad!"},
        {"msisdn": "962779234567", "message": "Hello Fatima!"}
      ],
      "message": "Default message",
      "encoding": "GSM7",
      "priority": "NORMAL",
      "dlr_url": "https://myapp.com/dlr",
      "campaign_name": "Summer Promo"
    }
    ```

    **Response**:
    ```json
    {
      "status": "SUCCESS",
      "campaign_id": "CAMP-123456789",
      "total_recipients": 2,
      "accepted": 2,
      "rejected": 0,
      "message": "Campaign created successfully"
    }
    ```
    """
    try:
        # Validate sender ID
        if hasattr(current_user, 'allowed_sender_ids'):
            allowed_senders = current_user.allowed_sender_ids or []
            if allowed_senders and request.sender not in allowed_senders:
                raise HTTPException(
                    status_code=status.HTTP_403_FORBIDDEN,
                    detail=f"Sender ID '{request.sender}' not allowed for this account"
                )

        # Check daily quota
        if hasattr(current_user, 'max_msg_per_day'):
            usage = unified_auth_service.get_daily_usage(db, current_user.id)
            if usage['remaining'] < len(request.recipients):
                raise HTTPException(
                    status_code=status.HTTP_429_TOO_MANY_REQUESTS,
                    detail=f"Daily quota exceeded. Remaining: {usage['remaining']}"
                )

        # Generate campaign ID
        campaign_id = f"CAMP-{secrets.token_hex(8).upper()}"

        # Create campaign
        campaign = Campaign(
            campaign_id=campaign_id,
            customer_id=current_user.customer_id if hasattr(current_user, 'customer_id') else None,
            user_id=current_user.id,
            name=request.campaign_name or f"API Campaign {datetime.utcnow().strftime('%Y%m%d-%H%M%S')}",
            description=f"API Bulk submission via HTTP API",
            submission_channel='HTTP_API',
            sender_id=request.sender,
            message_content=request.message,
            encoding=request.encoding,
            total_recipients=len(request.recipients),
            recipients_source='API',
            status='APPROVED' if not hasattr(current_user, 'requires_approval') else 'PENDING_APPROVAL',
            schedule_type='SCHEDULED' if request.schedule_time else 'IMMEDIATE',
            scheduled_time=request.schedule_time,
            priority=request.priority,
            dlr_required=True,
            dlr_callback_url=request.dlr_url,
            created_by=current_user.id
        )
        db.add(campaign)
        db.flush()

        # Create messages
        accepted = 0
        rejected = 0
        errors = []

        for recipient in request.recipients:
            try:
                # Validate MSISDN format
                if not recipient.msisdn.isdigit() or len(recipient.msisdn) < 10:
                    errors.append({
                        'msisdn': recipient.msisdn,
                        'error': 'Invalid phone number format'
                    })
                    rejected += 1
                    continue

                # Determine message content
                message_text = recipient.message or request.message
                if recipient.variables:
                    for key, value in recipient.variables.items():
                        message_text = message_text.replace(f"%{key}%", value)

                # Generate message ID
                message_id = f"MSG-{secrets.token_hex(12).upper()}"

                # Create message
                message = Message(
                    message_id=message_id,
                    customer_id=current_user.customer_id if hasattr(current_user, 'customer_id') else None,
                    user_id=current_user.id,
                    campaign_id=campaign.id,
                    from_addr=request.sender,
                    to_addr=recipient.msisdn,
                    message_text=message_text,
                    encoding=request.encoding,
                    submission_channel='HTTP_API',
                    status='QUEUED',
                    metadata={
                        'source': 'HTTP_API',
                        'priority': request.priority,
                        'dlr_url': request.dlr_url
                    }
                )
                db.add(message)
                accepted += 1

            except Exception as e:
                errors.append({
                    'msisdn': recipient.msisdn,
                    'error': str(e)
                })
                rejected += 1

        db.commit()

        # Update quota
        today = datetime.utcnow().date()
        quota = db.query(QuotaUsage).filter(
            QuotaUsage.user_id == current_user.id,
            QuotaUsage.period_date == today,
            QuotaUsage.period_type == 'DAILY'
        ).first()

        if quota:
            quota.messages_sent += accepted
            quota.updated_at = datetime.utcnow()
        else:
            quota = QuotaUsage(
                customer_id=current_user.customer_id if hasattr(current_user, 'customer_id') else None,
                user_id=current_user.id,
                period_date=today,
                period_type='DAILY',
                messages_sent=accepted
            )
            db.add(quota)

        db.commit()

        return SendBulkResponse(
            status="SUCCESS",
            campaign_id=campaign_id,
            total_recipients=len(request.recipients),
            accepted=accepted,
            rejected=rejected,
            message=f"Campaign created successfully. {accepted} messages queued, {rejected} rejected.",
            errors=errors if errors else None
        )

    except HTTPException:
        raise
    except Exception as e:
        db.rollback()
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Error processing bulk request: {str(e)}"
        )


@router.post("/getDLR", response_model=DLRQueryResponse)
async def get_delivery_reports(
    request: DLRQueryRequest,
    current_user: User = Depends(get_current_user_api_key),
    db: Session = Depends(get_db)
):
    """
    Query delivery reports

    **Authentication**: API Key (Bearer token)

    **Request Body**:
    ```json
    {
      "message_ids": ["MSG-ABC123", "MSG-DEF456"],
      "campaign_id": "CAMP-123456789",
      "from_date": "2025-01-01T00:00:00",
      "to_date": "2025-01-31T23:59:59",
      "limit": 100
    }
    ```
    """
    try:
        query = db.query(Message).filter(
            Message.user_id == current_user.id
        )

        # Apply filters
        if request.message_ids:
            query = query.filter(Message.message_id.in_(request.message_ids))

        if request.campaign_id:
            campaign = db.query(Campaign).filter(
                Campaign.campaign_id == request.campaign_id,
                Campaign.user_id == current_user.id
            ).first()
            if campaign:
                query = query.filter(Message.campaign_id == campaign.id)

        if request.from_date:
            query = query.filter(Message.submission_timestamp >= request.from_date)

        if request.to_date:
            query = query.filter(Message.submission_timestamp <= request.to_date)

        # Execute query
        messages = query.order_by(Message.submission_timestamp.desc()).limit(request.limit).all()

        # Build response
        records = []
        for msg in messages:
            records.append(DLRRecord(
                message_id=msg.message_id,
                msisdn=msg.to_addr,
                status=msg.status,
                dlr_status=msg.dlr_status,
                submitted_at=msg.submission_timestamp,
                delivered_at=msg.delivered_at,
                error_code=msg.error_code
            ))

        return DLRQueryResponse(
            status="SUCCESS",
            total=len(records),
            records=records
        )

    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Error querying DLR: {str(e)}"
        )


@router.get("/getBalance", response_model=BalanceResponse)
async def get_balance(
    current_user: User = Depends(get_current_user_api_key),
    db: Session = Depends(get_db)
):
    """
    Get account balance and quota information

    **Authentication**: API Key (Bearer token)

    **Response**:
    ```json
    {
      "status": "SUCCESS",
      "balance": 25000.50,
      "credit_limit": 50000.00,
      "currency": "JOD",
      "messages_sent_today": 12500,
      "messages_remaining_today": 487500,
      "daily_quota": 500000
    }
    ```
    """
    try:
        # Get customer balance
        balance = 0.0
        credit_limit = 0.0
        if hasattr(current_user, 'customer') and current_user.customer:
            balance = float(current_user.customer.balance)
            credit_limit = float(current_user.customer.credit_limit)

        # Get daily usage
        usage = unified_auth_service.get_daily_usage(db, current_user.id)

        return BalanceResponse(
            status="SUCCESS",
            balance=balance,
            credit_limit=credit_limit,
            currency="JOD",
            messages_sent_today=usage['messages_sent'],
            messages_remaining_today=usage['remaining'],
            daily_quota=usage['max_messages']
        )

    except Exception as e:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Error retrieving balance: {str(e)}"
        )


@router.post("/dlrCallback")
async def receive_dlr_callback(
    request: Request,
    db: Session = Depends(get_db)
):
    """
    Receive DLR callback from SMSC

    This endpoint receives delivery reports from SMSC and updates message status.
    It then forwards the DLR to the customer's callback URL if configured.
    """
    try:
        body = await request.json()

        message_id = body.get('message_id')
        dlr_status = body.get('status')
        error_code = body.get('error_code')

        if not message_id:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="message_id required"
            )

        # Find message
        message = db.query(Message).filter(
            Message.message_id == message_id
        ).first()

        if not message:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="Message not found"
            )

        # Update message status
        message.dlr_status = dlr_status
        message.dlr_timestamp = datetime.utcnow()
        message.error_code = error_code

        if dlr_status == 'DELIVERED':
            message.status = 'DELIVERED'
            message.delivered_at = datetime.utcnow()
        elif dlr_status in ['FAILED', 'REJECTED', 'EXPIRED']:
            message.status = 'FAILED'

        db.commit()

        # TODO: Forward DLR to customer callback URL
        # This will be implemented in the DLR handler service

        return {"status": "SUCCESS", "message": "DLR received"}

    except HTTPException:
        raise
    except Exception as e:
        db.rollback()
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Error processing DLR: {str(e)}"
        )
