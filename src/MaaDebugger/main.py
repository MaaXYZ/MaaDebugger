from nicegui import ui

from .webpage import index_page
from .webpage import reco_page

import socket
import argparse


def check_interface(interface: str, family: socket.AddressFamily,port: int):
    """
    Helper function to check a specific interface
    """
    try:
        with socket.socket(family, socket.SOCK_STREAM) as s:
            s.settimeout(1)
            return s.connect_ex((interface, port)) == 0
    except socket.error:
        return False

def is_port_in_use(port: int) -> bool:
    """
    Check if the port is in use
    """
    interfaces = [
        ("127.0.0.1", socket.AF_INET), 
        ("0.0.0.0", socket.AF_INET), 
        ("localhost", socket.AF_INET), 
        ("::1", socket.AF_INET6), 
        ("::", socket.AF_INET6)
    ]

    for interface, family in interfaces:
        if check_interface(interface, family, port):
            return True

    return False

def find_available_port(start_port: int) -> int:
    """
    Find the first available port starting from start_port
    """
    port = start_port
    while is_port_in_use(port):
        print(f"Port {port} is in use, trying next port...")
        port += 1
    print(f"Found available port: {port}")
    return port

def main():
    parser = argparse.ArgumentParser(description="A debugger specifically for MaaFramework.")
    parser.add_argument("--port", type=int, help="run port")

    args = parser.parse_args()
    specified_port: int | None = args.port

    if specified_port is not None:
        if is_port_in_use(specified_port):
            print(f"Specified port {specified_port} is in use.")
            return
        else:
            port = specified_port
    else:
        port = find_available_port(8011)

    ui.dark_mode()  # auto dark mode
    ui.run(
        port=port, title="Maa Debugger", storage_secret="maadbg", reload=False
    )  # , root_path="/proxy/8080")
