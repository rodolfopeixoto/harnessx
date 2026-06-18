from pydantic import BaseModel, Field


class Product(BaseModel):
    id: str
    name: str
    price_cents: int = Field(ge=0)


class CartItem(BaseModel):
    product_id: str
    quantity: int = Field(ge=1)


class AddToCart(BaseModel):
    product_id: str
    quantity: int = Field(ge=1, le=99)


class Cart(BaseModel):
    user_id: str
    items: list[CartItem] = Field(default_factory=list)
    total_cents: int = 0


class CheckoutRequest(BaseModel):
    user_id: str


class CheckoutResponse(BaseModel):
    order_id: str
    user_id: str
    total_cents: int
    items: list[CartItem]
