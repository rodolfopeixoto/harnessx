from fastapi import APIRouter, HTTPException

from app import storage
from app.models import AddToCart, Cart

router = APIRouter(prefix="/cart", tags=["cart"])


@router.get("/{user_id}", response_model=Cart)
def read_cart(user_id: str) -> Cart:
    return storage.get_cart(user_id)


@router.post("/{user_id}", response_model=Cart)
def add_item(user_id: str, body: AddToCart) -> Cart:
    try:
        return storage.add_to_cart(user_id, body.product_id, body.quantity)
    except KeyError as exc:
        raise HTTPException(status_code=404, detail=f"product {exc.args[0]} not found") from exc
