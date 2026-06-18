from fastapi import APIRouter, HTTPException

from app import storage
from app.models import Product

router = APIRouter(prefix="/products", tags=["products"])


@router.get("", response_model=list[Product])
def list_products() -> list[Product]:
    return storage.list_products()


@router.get("/{product_id}", response_model=Product)
def get_product(product_id: str) -> Product:
    product = storage.get_product(product_id)
    if product is None:
        raise HTTPException(status_code=404, detail="product not found")
    return product
