from __future__ import annotations

import uuid
from threading import RLock

from app.models import Cart, CartItem, Product

_PRODUCTS: dict[str, Product] = {
    "sku-001": Product(id="sku-001", name="Coffee Beans 500g", price_cents=1499),
    "sku-002": Product(id="sku-002", name="French Press", price_cents=2999),
    "sku-003": Product(id="sku-003", name="Espresso Cup", price_cents=899),
}

_CARTS: dict[str, Cart] = {}
_ORDERS: dict[str, dict] = {}
_LOCK = RLock()


def list_products() -> list[Product]:
    with _LOCK:
        return list(_PRODUCTS.values())


def get_product(product_id: str) -> Product | None:
    with _LOCK:
        return _PRODUCTS.get(product_id)


def get_cart(user_id: str) -> Cart:
    with _LOCK:
        return _CARTS.get(user_id) or Cart(user_id=user_id)


def add_to_cart(user_id: str, product_id: str, quantity: int) -> Cart:
    with _LOCK:
        product = _PRODUCTS.get(product_id)
        if product is None:
            raise KeyError(product_id)
        cart = _CARTS.get(user_id) or Cart(user_id=user_id)
        for item in cart.items:
            if item.product_id == product_id:
                item.quantity += quantity
                break
        else:
            cart.items.append(CartItem(product_id=product_id, quantity=quantity))
        cart.total_cents = sum(_PRODUCTS[i.product_id].price_cents * i.quantity for i in cart.items)
        _CARTS[user_id] = cart
        return cart


def checkout(user_id: str) -> dict:
    with _LOCK:
        cart = _CARTS.get(user_id)
        if cart is None or not cart.items:
            raise ValueError("cart empty")
        order_id = f"ord-{uuid.uuid4().hex[:8]}"
        order = {
            "order_id": order_id,
            "user_id": user_id,
            "total_cents": cart.total_cents,
            "items": [i.model_dump() for i in cart.items],
        }
        _ORDERS[order_id] = order
        _CARTS.pop(user_id, None)
        return order


def reset_for_tests() -> None:
    with _LOCK:
        _CARTS.clear()
        _ORDERS.clear()
