from .main import run

from .utils import args

if __name__ == "__main__":
    host = args.get_host()
    port = args.get_port() or 8011
    show = args.get_show()
    dark = args.get_dark()

    run(
        host=host,
        port=port,
        show=show,
        dark=dark,
    )
