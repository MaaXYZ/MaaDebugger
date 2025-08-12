from pathlib import Path

ASSETS_PATH = Path(__file__).parent

FAVICON_PATH = (
    ASSETS_PATH / "favicon.png" if (ASSETS_PATH / "favicon.png").exists() else None
)
