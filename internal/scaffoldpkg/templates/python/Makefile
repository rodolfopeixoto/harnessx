.PHONY: setup lint test run clean

setup:
	python3 -m venv .venv
	.venv/bin/pip install -r requirements.txt

lint:
	.venv/bin/ruff check .

test:
	.venv/bin/pytest -q

run:
	.venv/bin/uvicorn app:app --reload

clean:
	rm -rf .venv .pytest_cache __pycache__ */__pycache__
