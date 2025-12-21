from pydantic import BaseModel, Field
from typing import List, Optional
from uuid import UUID
from datetime import datetime

# --- Общие модели ---
class PrivilegeShortInfo(BaseModel):
    balance: int
    status: str # BRONZE, SILVER, GOLD

# --- Flight DTO ---
class FlightResponse(BaseModel):
    flightNumber: str
    fromAirport: str
    toAirport: str
    date: str
    price: int

class PaginationResponse(BaseModel):
    page: int
    pageSize: int
    totalElements: int
    items: List[FlightResponse]

# --- Ticket DTO ---
class TicketResponse(BaseModel):
    ticketUid: UUID
    flightNumber: str
    fromAirport: str
    toAirport: str
    date: str
    price: int
    status: str

# --- User Info DTO ---
class UserInfoResponse(BaseModel):
    tickets: List[TicketResponse]
    privilege: PrivilegeShortInfo

# --- Purchase DTO ---
class TicketPurchaseRequest(BaseModel):
    flightNumber: str
    price: int
    paidFromBalance: bool

class TicketPurchaseResponse(TicketResponse):
    paidByMoney: int
    paidByBonuses: int
    privilege: PrivilegeShortInfo