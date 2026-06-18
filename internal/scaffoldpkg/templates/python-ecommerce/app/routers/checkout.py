from fastapi import APIRouter, HTTPException

from app import storage
from app.models import CheckoutRequest, CheckoutResponse

router = APIRouter(prefix="/checkout", tags=["checkout"])


@router.post("", response_model=CheckoutResponse)
def checkout(body: CheckoutRequest) -> CheckoutResponse:
    try:
        order = storage.checkout(body.user_id)
    except ValueError as exc:
        raise HTTPException(status_code=400, detail=str(exc)) from exc
    return CheckoutResponse(**order)
