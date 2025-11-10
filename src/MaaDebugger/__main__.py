from . import MaaDebugger
from .utils.arg_parser import ArgParser


def main():
    args = ArgParser()
    host = args.get_host()
    port = args.get_port()
    show = args.get_show()
    dark = args.get_dark()

    MaaDebugger.run(
        host=host,
        port=port,
        show=show,
        dark=dark,
    )


if __name__ == "__main__":
    main()
