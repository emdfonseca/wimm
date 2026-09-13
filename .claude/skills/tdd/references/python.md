# Python

## pytest, in `tests/`

`tests/test_<module>.py`, functions named `test_<behaviour_as_sentence>`:

```python
def test_parse_amount_rejects_negative():
    with pytest.raises(ValueError):
        parse_amount("-1.00")

@pytest.mark.parametrize("raw, cents", [("10.50", 1050), ("0", 0), ("1", 100)])
def test_parse_amount_converts_to_cents(raw, cents):
    assert parse_amount(raw) == cents
```

Run one while cycling: `uv run pytest -k rejects_negative`.

## Fixtures over setup

Small `pytest` fixtures for fakes (an in-memory store, a fixed clock), composed rather than one giant `conftest.py`. A fixture that takes more than a few lines is a fake worth its own module under `tests/fakes/`.

## Property tests for the arithmetic

Parsers, money, date handling: after the example-based cycle establishes the shape, one Hypothesis property (`round-trips`, `never negative`, `associative`) often finds the case the examples missed. Add it as its own cycle — red is Hypothesis printing the counterexample.

## Types are a test

`mypy` / `pyright` in `just check` catches a class of bugs before any test runs. Write the signature first; a function whose types are wrong is red before the test is.
