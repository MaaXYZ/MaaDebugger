import argparse

from ..port_checker import PortChecker


class ArgParser:
    def __init__(self) -> None:
        self.parser = argparse.ArgumentParser(description="A debugger specifically for MaaFramework.")
        self._add_argument()
        self.args = self.parser.parse_args()

    def _add_argument(self):
        """
        Add command line arguments to the parser.
        """
        self.parser.add_argument("--port", type=int, help="run port")

    
    def get_port(self) -> int:
        """
        Determine the port to use based on the provided arguments.
        """
        specified_port: int | None = self.args.port

        if specified_port is not None:
            if PortChecker.is_port_in_use(specified_port):
                print(f"Specified port {specified_port} is in use.")
                return -1
            else:
                port = specified_port
        else:
            port = PortChecker.find_available_port(8011)
        return port