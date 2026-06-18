def test_checkout_empty_cart_returns_400(client) -> None:
    res = client.post("/checkout", json={"user_id": "alice"})
    assert res.status_code == 400


def test_checkout_success_clears_cart(client) -> None:
    client.post("/cart/alice", json={"product_id": "sku-002", "quantity": 1})
    res = client.post("/checkout", json={"user_id": "alice"})
    assert res.status_code == 200
    order = res.json()
    assert order["user_id"] == "alice"
    assert order["total_cents"] == 2999
    cart_after = client.get("/cart/alice").json()
    assert cart_after["items"] == []
