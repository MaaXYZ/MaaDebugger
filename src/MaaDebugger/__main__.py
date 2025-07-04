from . import MaaDebugger
from .utils import args

if __name__ == "__main__":
    args.parse()

    host = args.get_host()
    port = args.get_port() or 8011
    show = args.get_hide()
    dark = args.get_dark()

    MaaDebugger.run(
        host=host,
        port=port,
        show=show,
        dark=dark,
    )
