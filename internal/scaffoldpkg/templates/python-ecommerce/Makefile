.PHONY: test lint run

test:
	.venv/bin/pytest -q

lint:
	.venv/bin/ruff check .

run:
	.venv/bin/uvicorn app.main:app --reload
