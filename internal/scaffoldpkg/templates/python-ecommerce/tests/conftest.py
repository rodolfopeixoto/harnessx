import pytest
from fastapi.testclient import TestClient

from app import storage
from app.main import app


@pytest.fixture()
def client() -> TestClient:
    storage.reset_for_tests()
    return TestClient(app)
