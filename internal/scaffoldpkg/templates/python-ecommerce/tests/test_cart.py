def test_empty_cart_returns_zero_total(client) -> None:
    res = client.get("/cart/alice")
    assert res.status_code == 200
    cart = res.json()
    assert cart["user_id"] == "alice"
    assert cart["items"] == []
    assert cart["total_cents"] == 0


def test_add_to_cart_accumulates_total(client) -> None:
    res = client.post("/cart/alice", json={"product_id": "sku-001", "quantity": 2})
    assert res.status_code == 200
    body = res.json()
    assert body["total_cents"] == 1499 * 2
    assert body["items"][0]["quantity"] == 2


def test_add_to_cart_missing_product_returns_404(client) -> None:
    res = client.post("/cart/alice", json={"product_id": "sku-ghost", "quantity": 1})
    assert res.status_code == 404
