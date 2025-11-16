#!/usr/bin/env python3
"""
Routing Engine Service
Dynamic SMSC routing based on rules and conditions
"""

from typing import Optional, Dict, Any, List, Tuple
from datetime import datetime, time
import re
import logging
from sqlalchemy.orm import Session
from sqlalchemy import and_, or_, func

from src.models.routing import SMSCConnection, RoutingRule, RoutingLog, CountryCode
from src.models.campaign import Message

logger = logging.getLogger(__name__)


class RoutingEngine:
    """
    Routing Engine for dynamic SMSC selection

    Matches messages against routing rules based on:
    - MSISDN prefix
    - Sender ID
    - Customer ID
    - Message type
    - Country code
    - Custom regex patterns
    - Combined conditions
    """

    def __init__(self):
        self.cache_ttl = 300  # 5 minutes cache for rules

    def route_message(
        self,
        db: Session,
        msisdn: str,
        sender_id: str = None,
        message_type: str = None,
        customer_id: int = None,
        user_id: int = None,
        message_id: int = None,
        campaign_id: int = None
    ) -> Tuple[Optional[SMSCConnection], Optional[RoutingRule], bool]:
        """
        Route a message to appropriate SMSC

        Args:
            db: Database session
            msisdn: Recipient phone number
            sender_id: Sender ID
            message_type: Message type (OTP, PROMO, etc.)
            customer_id: Customer ID
            user_id: User ID
            message_id: Message ID for logging
            campaign_id: Campaign ID for logging

        Returns:
            Tuple: (selected_smsc, routing_rule, is_fallback)
        """
        start_time = datetime.utcnow()

        try:
            # Find matching routing rule
            rule = self._find_matching_rule(
                db, msisdn, sender_id, message_type, customer_id, user_id
            )

            if not rule:
                # Use default route
                logger.info(f"No specific rule found for {msisdn}, using default route")
                smsc = self._get_default_smsc(db)
                is_fallback = False
            else:
                # Get primary SMSC
                smsc = db.query(SMSCConnection).filter(
                    SMSCConnection.smsc_id == rule.smsc_id
                ).first()

                # Check if SMSC is available
                if not smsc or not self._is_smsc_available(smsc):
                    logger.warning(f"Primary SMSC {rule.smsc_id} not available, trying fallback")

                    # Try fallback SMSC
                    if rule.fallback_smsc_id:
                        smsc = db.query(SMSCConnection).filter(
                            SMSCConnection.smsc_id == rule.fallback_smsc_id
                        ).first()
                        is_fallback = True
                    else:
                        smsc = self._get_default_smsc(db)
                        is_fallback = True
                else:
                    is_fallback = False

            # Log routing decision
            routing_time_ms = int((datetime.utcnow() - start_time).total_seconds() * 1000)

            if smsc:
                self._log_routing_decision(
                    db=db,
                    message_id=message_id,
                    campaign_id=campaign_id,
                    msisdn=msisdn,
                    sender_id=sender_id,
                    message_type=message_type,
                    rule=rule,
                    smsc=smsc,
                    is_fallback=is_fallback,
                    routing_time_ms=routing_time_ms,
                    status='SUCCESS' if not is_fallback else 'FALLBACK'
                )

                logger.info(f"Routed {msisdn} to SMSC {smsc.smsc_code} via rule {rule.rule_code if rule else 'DEFAULT'}")
                return smsc, rule, is_fallback
            else:
                # No route found
                self._log_routing_decision(
                    db=db,
                    message_id=message_id,
                    campaign_id=campaign_id,
                    msisdn=msisdn,
                    sender_id=sender_id,
                    message_type=message_type,
                    rule=None,
                    smsc=None,
                    is_fallback=False,
                    routing_time_ms=routing_time_ms,
                    status='NO_ROUTE'
                )

                logger.error(f"No route found for {msisdn}")
                return None, None, False

        except Exception as e:
            logger.error(f"Error routing message: {e}", exc_info=True)
            return None, None, False

    def _find_matching_rule(
        self,
        db: Session,
        msisdn: str,
        sender_id: str = None,
        message_type: str = None,
        customer_id: int = None,
        user_id: int = None
    ) -> Optional[RoutingRule]:
        """
        Find the first matching routing rule based on priority

        Args:
            db: Database session
            msisdn: Recipient phone number
            sender_id: Sender ID
            message_type: Message type
            customer_id: Customer ID
            user_id: User ID

        Returns:
            RoutingRule: Matching rule or None
        """
        # Get all active rules ordered by priority
        rules = db.query(RoutingRule).filter(
            RoutingRule.is_active == True
        ).filter(
            or_(
                RoutingRule.customer_id == customer_id,
                RoutingRule.customer_id.is_(None)
            )
        ).order_by(RoutingRule.priority.asc()).all()

        current_time = datetime.utcnow().time()

        for rule in rules:
            # Check time-based routing
            if rule.enable_time_based:
                if not self._is_within_active_hours(rule, current_time):
                    continue

            # Match based on condition type
            if self._matches_condition(rule, msisdn, sender_id, message_type, customer_id):
                return rule

        return None

    def _matches_condition(
        self,
        rule: RoutingRule,
        msisdn: str,
        sender_id: str = None,
        message_type: str = None,
        customer_id: int = None
    ) -> bool:
        """
        Check if message matches rule condition

        Args:
            rule: Routing rule
            msisdn: Recipient phone number
            sender_id: Sender ID
            message_type: Message type
            customer_id: Customer ID

        Returns:
            bool: True if matches
        """
        if rule.condition_type == 'PREFIX':
            # Match MSISDN prefix
            if rule.msisdn_prefix:
                return msisdn.startswith(rule.msisdn_prefix)
            else:
                return msisdn.startswith(rule.condition_value)

        elif rule.condition_type == 'SENDER':
            # Match sender ID pattern
            if sender_id and rule.sender_id_pattern:
                pattern = re.compile(rule.sender_id_pattern)
                return bool(pattern.match(sender_id))
            return False

        elif rule.condition_type == 'CUSTOMER':
            # Match customer ID
            return rule.customer_id == customer_id

        elif rule.condition_type == 'MESSAGE_TYPE':
            # Match message type
            return rule.message_type == message_type

        elif rule.condition_type == 'COUNTRY':
            # Match country code
            country_code = self._extract_country_code(msisdn)
            return country_code == rule.country_code

        elif rule.condition_type == 'REGEX':
            # Match regex pattern
            if rule.regex_pattern:
                pattern = re.compile(rule.regex_pattern)
                return bool(pattern.match(msisdn))
            return False

        elif rule.condition_type == 'COMBINED':
            # Match combined conditions (all must match)
            conditions = rule.combined_conditions or {}
            matches = []

            if 'prefix' in conditions:
                matches.append(msisdn.startswith(conditions['prefix']))

            if 'sender' in conditions and sender_id:
                matches.append(sender_id == conditions['sender'])

            if 'message_type' in conditions and message_type:
                matches.append(message_type == conditions['message_type'])

            if 'country' in conditions:
                country_code = self._extract_country_code(msisdn)
                matches.append(country_code == conditions['country'])

            return all(matches) if matches else False

        return False

    def _extract_country_code(self, msisdn: str) -> Optional[str]:
        """
        Extract country code from MSISDN

        Args:
            msisdn: Phone number

        Returns:
            str: Country code or None
        """
        # Common country code lengths: 1-4 digits
        for length in [4, 3, 2, 1]:
            prefix = msisdn[:length]
            if prefix.isdigit():
                return prefix
        return None

    def _is_within_active_hours(self, rule: RoutingRule, current_time: time) -> bool:
        """
        Check if current time is within rule's active hours

        Args:
            rule: Routing rule
            current_time: Current time

        Returns:
            bool: True if within active hours
        """
        if not rule.active_hours_start or not rule.active_hours_end:
            return True

        if rule.active_hours_start <= rule.active_hours_end:
            # Normal range (e.g., 09:00 to 17:00)
            return rule.active_hours_start <= current_time <= rule.active_hours_end
        else:
            # Overnight range (e.g., 22:00 to 06:00)
            return current_time >= rule.active_hours_start or current_time <= rule.active_hours_end

    def _is_smsc_available(self, smsc: SMSCConnection) -> bool:
        """
        Check if SMSC is available for routing

        Args:
            smsc: SMSC connection

        Returns:
            bool: True if available
        """
        if not smsc:
            return False

        # Check status
        if smsc.status != 'CONNECTED':
            return False

        # Check route mode
        if smsc.route_mode not in ['ACTIVE', 'STANDBY']:
            return False

        # Check error rate threshold
        if smsc.delivery_rate > 0 and smsc.delivery_rate < (100 - smsc.error_rate_threshold):
            logger.warning(f"SMSC {smsc.smsc_code} delivery rate {smsc.delivery_rate}% below threshold")
            return False

        return True

    def _get_default_smsc(self, db: Session) -> Optional[SMSCConnection]:
        """
        Get default SMSC for routing

        Args:
            db: Database session

        Returns:
            SMSCConnection: Default SMSC or None
        """
        return db.query(SMSCConnection).filter(
            SMSCConnection.is_default_route == True,
            SMSCConnection.route_mode == 'ACTIVE',
            SMSCConnection.status == 'CONNECTED'
        ).order_by(SMSCConnection.priority.asc()).first()

    def _log_routing_decision(
        self,
        db: Session,
        message_id: int = None,
        campaign_id: int = None,
        msisdn: str = None,
        sender_id: str = None,
        message_type: str = None,
        rule: RoutingRule = None,
        smsc: SMSCConnection = None,
        is_fallback: bool = False,
        routing_time_ms: int = 0,
        status: str = 'SUCCESS'
    ):
        """
        Log routing decision to database

        Args:
            db: Database session
            message_id: Message ID
            campaign_id: Campaign ID
            msisdn: Recipient phone number
            sender_id: Sender ID
            message_type: Message type
            rule: Routing rule used
            smsc: Selected SMSC
            is_fallback: Whether fallback was used
            routing_time_ms: Routing time in milliseconds
            status: Routing status
        """
        try:
            log = RoutingLog(
                message_id=message_id,
                campaign_id=campaign_id,
                msisdn=msisdn,
                sender_id=sender_id,
                message_type=message_type,
                rule_id=rule.rule_id if rule else None,
                rule_name=rule.rule_name if rule else None,
                selected_smsc_id=smsc.smsc_id if smsc else None,
                smsc_name=smsc.smsc_name if smsc else None,
                is_fallback=is_fallback,
                routing_status=status,
                routing_time_ms=routing_time_ms
            )
            db.add(log)
            db.commit()

        except Exception as e:
            logger.error(f"Error logging routing decision: {e}", exc_info=True)
            db.rollback()

    def get_routing_statistics(
        self,
        db: Session,
        customer_id: int = None,
        smsc_id: int = None,
        from_date: datetime = None,
        to_date: datetime = None
    ) -> Dict[str, Any]:
        """
        Get routing statistics

        Args:
            db: Database session
            customer_id: Filter by customer
            smsc_id: Filter by SMSC
            from_date: Start date
            to_date: End date

        Returns:
            Dict: Statistics
        """
        query = db.query(RoutingLog)

        if customer_id:
            # Filter by customer's campaigns
            from src.models.campaign import Campaign
            query = query.join(Campaign).filter(Campaign.customer_id == customer_id)

        if smsc_id:
            query = query.filter(RoutingLog.selected_smsc_id == smsc_id)

        if from_date:
            query = query.filter(RoutingLog.created_at >= from_date)

        if to_date:
            query = query.filter(RoutingLog.created_at <= to_date)

        total = query.count()
        success = query.filter(RoutingLog.routing_status == 'SUCCESS').count()
        fallback = query.filter(RoutingLog.routing_status == 'FALLBACK').count()
        failed = query.filter(RoutingLog.routing_status == 'FAILED').count()
        no_route = query.filter(RoutingLog.routing_status == 'NO_ROUTE').count()

        # Average routing time
        avg_time_result = query.with_entities(
            func.avg(RoutingLog.routing_time_ms)
        ).scalar()
        avg_routing_time = float(avg_time_result) if avg_time_result else 0.0

        return {
            'total_routes': total,
            'successful': success,
            'fallback_used': fallback,
            'failed': failed,
            'no_route': no_route,
            'success_rate': round((success / total * 100), 2) if total > 0 else 0.0,
            'fallback_rate': round((fallback / total * 100), 2) if total > 0 else 0.0,
            'avg_routing_time_ms': round(avg_routing_time, 2)
        }

    def test_route(
        self,
        db: Session,
        msisdn: str,
        sender_id: str = None,
        message_type: str = None,
        customer_id: int = None
    ) -> Dict[str, Any]:
        """
        Test routing for a given MSISDN (without logging)

        Args:
            db: Database session
            msisdn: Phone number to test
            sender_id: Sender ID
            message_type: Message type
            customer_id: Customer ID

        Returns:
            Dict: Test result with matched rule and SMSC
        """
        rule = self._find_matching_rule(db, msisdn, sender_id, message_type, customer_id)

        if rule:
            smsc = db.query(SMSCConnection).filter(
                SMSCConnection.smsc_id == rule.smsc_id
            ).first()

            return {
                'matched': True,
                'rule': {
                    'rule_id': rule.rule_id,
                    'rule_code': rule.rule_code,
                    'rule_name': rule.rule_name,
                    'condition_type': rule.condition_type,
                    'priority': rule.priority
                },
                'smsc': {
                    'smsc_id': smsc.smsc_id if smsc else None,
                    'smsc_code': smsc.smsc_code if smsc else None,
                    'smsc_name': smsc.smsc_name if smsc else None,
                    'status': smsc.status if smsc else None,
                    'is_available': self._is_smsc_available(smsc) if smsc else False
                },
                'fallback_smsc': {
                    'smsc_id': rule.fallback_smsc_id
                } if rule.fallback_smsc_id else None
            }
        else:
            # Would use default route
            default_smsc = self._get_default_smsc(db)

            return {
                'matched': False,
                'rule': None,
                'smsc': {
                    'smsc_id': default_smsc.smsc_id if default_smsc else None,
                    'smsc_code': default_smsc.smsc_code if default_smsc else None,
                    'smsc_name': default_smsc.smsc_name if default_smsc else None,
                    'status': default_smsc.status if default_smsc else None,
                    'is_default': True
                } if default_smsc else None
            }


# Global routing engine instance
routing_engine = RoutingEngine()
