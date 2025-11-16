#!/usr/bin/env python3
"""
DLR (Delivery Report) Handler
Processes delivery reports from SMSC and forwards to customer callbacks
"""

from typing import Optional, Dict, Any
from datetime import datetime
import asyncio
import aiohttp
from sqlalchemy.orm import Session
from sqlalchemy import and_

from src.models.campaign import Message
from src.models.user import User
import logging

logger = logging.getLogger(__name__)


class DLRHandler:
    """
    DLR Handler Service
    Processes delivery reports and forwards to customer callback URLs
    """

    def __init__(self):
        self.callback_timeout = 10  # seconds
        self.max_retries = 3

    async def process_dlr(
        self,
        db: Session,
        message_id: str = None,
        smpp_msg_id: str = None,
        msisdn: str = None,
        dlr_status: str = None,
        error_code: str = None,
        dlr_text: str = None,
        submit_time: datetime = None,
        done_time: datetime = None,
        smsc_id: str = None
    ) -> bool:
        """
        Process delivery report

        Args:
            db: Database session
            message_id: Internal message ID
            smpp_msg_id: SMPP message ID from SMSC
            msisdn: Recipient phone number
            dlr_status: DLR status (DELIVERED, FAILED, REJECTED, EXPIRED, etc.)
            error_code: Error code if failed
            dlr_text: DLR text description
            submit_time: Message submission time
            done_time: DLR received time
            smsc_id: SMSC identifier

        Returns:
            bool: True if processed successfully
        """
        try:
            # Find message by ID or SMPP message ID
            message = None
            if message_id:
                message = db.query(Message).filter(
                    Message.message_id == message_id
                ).first()
            elif smpp_msg_id:
                message = db.query(Message).filter(
                    Message.smpp_msg_id == smpp_msg_id
                ).first()

            if not message:
                logger.warning(f"Message not found for DLR: message_id={message_id}, smpp_msg_id={smpp_msg_id}")
                return False

            # Update message with DLR
            message.dlr_status = dlr_status
            message.dlr_timestamp = datetime.utcnow()
            message.dlr_text = dlr_text
            message.error_code = error_code

            # Update message status
            if dlr_status == 'DELIVERED' or dlr_status == 'DELIVRD':
                message.status = 'DELIVERED'
                message.delivered_at = done_time or datetime.utcnow()
            elif dlr_status in ['FAILED', 'REJECTED', 'EXPIRED', 'UNDELIV']:
                message.status = 'FAILED'
            elif dlr_status in ['ACCEPTED', 'ENROUTE']:
                message.status = 'SENT'

            db.commit()

            logger.info(f"DLR processed: message_id={message.message_id}, status={dlr_status}")

            # Get campaign and check for callback URL
            if message.campaign and message.campaign.dlr_callback_url:
                await self._forward_dlr_to_callback(
                    db=db,
                    message=message,
                    callback_url=message.campaign.dlr_callback_url
                )

            # Update quota statistics
            await self._update_delivery_statistics(db, message)

            return True

        except Exception as e:
            logger.error(f"Error processing DLR: {e}", exc_info=True)
            db.rollback()
            return False

    async def _forward_dlr_to_callback(
        self,
        db: Session,
        message: Message,
        callback_url: str
    ) -> bool:
        """
        Forward DLR to customer callback URL

        Args:
            db: Database session
            message: Message object
            callback_url: Customer callback URL

        Returns:
            bool: True if callback successful
        """
        try:
            # Prepare callback payload
            payload = {
                'message_id': message.message_id,
                'msisdn': message.to_addr,
                'sender': message.from_addr,
                'status': message.dlr_status,
                'error_code': message.error_code,
                'dlr_text': message.dlr_text,
                'submitted_at': message.submission_timestamp.isoformat() if message.submission_timestamp else None,
                'delivered_at': message.delivered_at.isoformat() if message.delivered_at else None,
                'timestamp': datetime.utcnow().isoformat()
            }

            # Send HTTP POST to callback URL
            async with aiohttp.ClientSession() as session:
                async with session.post(
                    callback_url,
                    json=payload,
                    timeout=aiohttp.ClientTimeout(total=self.callback_timeout)
                ) as response:
                    response_text = await response.text()

                    if response.status == 200:
                        logger.info(f"DLR callback successful: message_id={message.message_id}, url={callback_url}")

                        # Update message metadata
                        metadata = message.metadata or {}
                        metadata['dlr_callback_sent'] = True
                        metadata['dlr_callback_at'] = datetime.utcnow().isoformat()
                        metadata['dlr_callback_response'] = response_text[:500]
                        message.metadata = metadata
                        db.commit()

                        return True
                    else:
                        logger.warning(f"DLR callback failed: message_id={message.message_id}, status={response.status}, response={response_text}")
                        return False

        except asyncio.TimeoutError:
            logger.error(f"DLR callback timeout: message_id={message.message_id}, url={callback_url}")
            return False
        except Exception as e:
            logger.error(f"Error forwarding DLR callback: {e}", exc_info=True)
            return False

    async def _update_delivery_statistics(
        self,
        db: Session,
        message: Message
    ) -> bool:
        """
        Update delivery statistics in quota table

        Args:
            db: Database session
            message: Message object

        Returns:
            bool: True if updated
        """
        try:
            from src.models.quota import QuotaUsage

            today = datetime.utcnow().date()

            # Find or create quota record
            quota = db.query(QuotaUsage).filter(
                and_(
                    QuotaUsage.user_id == message.user_id,
                    QuotaUsage.period_date == today,
                    QuotaUsage.period_type == 'DAILY'
                )
            ).first()

            if quota:
                if message.status == 'DELIVERED':
                    quota.messages_delivered += 1
                elif message.status == 'FAILED':
                    quota.messages_failed += 1

                quota.updated_at = datetime.utcnow()
                db.commit()

            return True

        except Exception as e:
            logger.error(f"Error updating delivery statistics: {e}", exc_info=True)
            db.rollback()
            return False

    def process_dlr_sync(
        self,
        db: Session,
        **kwargs
    ) -> bool:
        """
        Synchronous wrapper for process_dlr

        Args:
            db: Database session
            **kwargs: DLR parameters

        Returns:
            bool: True if processed successfully
        """
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        try:
            result = loop.run_until_complete(self.process_dlr(db, **kwargs))
            return result
        finally:
            loop.close()

    async def batch_process_dlrs(
        self,
        db: Session,
        dlr_list: list
    ) -> Dict[str, Any]:
        """
        Process multiple DLRs in batch

        Args:
            db: Database session
            dlr_list: List of DLR dictionaries

        Returns:
            Dict with processing statistics
        """
        processed = 0
        failed = 0

        for dlr in dlr_list:
            success = await self.process_dlr(db, **dlr)
            if success:
                processed += 1
            else:
                failed += 1

        return {
            'total': len(dlr_list),
            'processed': processed,
            'failed': failed
        }

    def get_message_dlr_status(
        self,
        db: Session,
        message_id: str
    ) -> Optional[Dict[str, Any]]:
        """
        Get DLR status for a specific message

        Args:
            db: Database session
            message_id: Message ID

        Returns:
            Dict with DLR status information
        """
        message = db.query(Message).filter(
            Message.message_id == message_id
        ).first()

        if not message:
            return None

        return {
            'message_id': message.message_id,
            'msisdn': message.to_addr,
            'status': message.status,
            'dlr_status': message.dlr_status,
            'error_code': message.error_code,
            'dlr_text': message.dlr_text,
            'submitted_at': message.submission_timestamp.isoformat() if message.submission_timestamp else None,
            'delivered_at': message.delivered_at.isoformat() if message.delivered_at else None,
            'dlr_timestamp': message.dlr_timestamp.isoformat() if message.dlr_timestamp else None
        }

    def get_campaign_dlr_summary(
        self,
        db: Session,
        campaign_id: int
    ) -> Dict[str, Any]:
        """
        Get DLR summary for entire campaign

        Args:
            db: Database session
            campaign_id: Campaign ID

        Returns:
            Dict with campaign DLR summary
        """
        from sqlalchemy import func

        # Count messages by status
        stats = db.query(
            Message.status,
            func.count(Message.id).label('count')
        ).filter(
            Message.campaign_id == campaign_id
        ).group_by(Message.status).all()

        summary = {
            'campaign_id': campaign_id,
            'total': 0,
            'pending': 0,
            'queued': 0,
            'sent': 0,
            'delivered': 0,
            'failed': 0,
            'rejected': 0,
            'expired': 0
        }

        for status, count in stats:
            summary['total'] += count
            summary[status.lower()] = count

        # Calculate percentages
        if summary['total'] > 0:
            summary['delivery_rate'] = round((summary['delivered'] / summary['total']) * 100, 2)
            summary['failure_rate'] = round((summary['failed'] / summary['total']) * 100, 2)
        else:
            summary['delivery_rate'] = 0.0
            summary['failure_rate'] = 0.0

        return summary


# Global DLR handler instance
dlr_handler = DLRHandler()
