from fastapi import FastAPI, Header, HTTPException
from .database import get_privilege_with_history, process_bonus_operation
from .schemas import PrivilegeInfoResponse, BonusOperationRequest, BonusOperationResponse

app = FastAPI(title="Bonus Service")

@app.get("/api/v1/privilege", response_model=PrivilegeInfoResponse)
async def get_privilege(x_user_name: str = Header(...)):
    data = get_privilege_with_history(x_user_name)
    if not data:
        raise HTTPException(status_code=404, detail="Privilege not found")
    return data

@app.post("/api/v1/privilege/calculate", response_model=BonusOperationResponse)
async def calculate_bonus(request: BonusOperationRequest, x_user_name: str = Header(...)):
    return process_bonus_operation(
        x_user_name, str(request.ticketUid), request.price, request.paidFromBalance
    )