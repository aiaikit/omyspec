"""Module entry point for `python -m specify_cli`.

The npm distribution launches the CLI this way (PYTHONPATH pointing at an
unpacked runtime directory) instead of relying on the console-script shim that
only exists for pip/uv installs.
"""
from . import main

if __name__ == "__main__":
    # Click builds the program name from ``__package__`` plus the executed
    # filename, so `-m specify_cli` would advertise `python -m specify_cli` in
    # usage and error text. Name the command the way pip/uv installs do.
    main(prog_name="specify")
