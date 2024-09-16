import socket


class PortChecker:
    @staticmethod
    def check_interface(interface: str, family: socket.AddressFamily,port: int):
        """
        Helper function to check a specific interface.
        """
        try:
            with socket.socket(family, socket.SOCK_STREAM) as s:
                s.settimeout(1)
                return s.connect_ex((interface, port)) == 0
        except socket.error:
            return False

    @staticmethod

    def is_port_in_use(port: int) -> bool:
        """
        Check if the port is in use.
        """
        interfaces = [
            ("127.0.0.1", socket.AF_INET), 
            ("0.0.0.0", socket.AF_INET), 
            ("localhost", socket.AF_INET), 
            ("::1", socket.AF_INET6), 
            ("::", socket.AF_INET6)
        ]

        for interface, family in interfaces:
            if PortChecker.check_interface(interface, family, port):
                return True

        return False

    @staticmethod
    def find_available_port(start_port: int) -> int:
        """
        Find the first available port starting from start_port.
        """
        port = start_port
        while PortChecker.is_port_in_use(port):
            print(f"Port {port} is in use, trying next port...")
            port += 1
        print(f"Found available port: {port}")
        return port