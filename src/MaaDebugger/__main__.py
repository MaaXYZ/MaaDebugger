from . import MaaDebugger
from .utils.arg_parser import ArgParser


def main():
    host = ArgParser.get_host()
    port = ArgParser.get_port()
    show = ArgParser.get_show()
    dark = ArgParser.get_dark()

    MaaDebugger.run(
        host=host,
        port=port,
        show=show,
        dark=dark,
    )


if __name__ == "__main__":
    main()
