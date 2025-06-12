from ...webpage.components.status_indicator import Status


class GlobalStatus:
    ctrl_connecting: Status = Status.PENDING
    ctrl_detecting: Status = Status.PENDING  # not required
    res_loading: Status = Status.PENDING
    task_running: Status = Status.PENDING
    agent_connecting: Status = Status.PENDING
