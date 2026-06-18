def test_list_products_has_three_defaults(client) -> None:
    res = client.get("/products")
    assert res.status_code == 200
    items = res.json()
    assert len(items) == 3
    ids = {item["id"] for item in items}
    assert {"sku-001", "sku-002", "sku-003"}.issubset(ids)


def test_get_product_known_sku(client) -> None:
    res = client.get("/products/sku-001")
    assert res.status_code == 200
    assert res.json()["name"] == "Coffee Beans 500g"


def test_get_product_unknown_sku_returns_404(client) -> None:
    res = client.get("/products/sku-missing")
    assert res.status_code == 404
