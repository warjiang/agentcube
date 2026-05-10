"""
KernelHub Python SDK
"""

import requests
from typing import Optional, Dict, Any


class KernelHubClient:
    """KernelHub Python SDK"""

    def __init__(self, base_url: str, auth_token: str):
        self.base_url = base_url.rstrip('/')
        self.auth_token = auth_token

    def _headers(self):
        return {
            'Authorization': f'Bearer {self.auth_token}',
            'Content-Type': 'application/json'
        }

    def create_session(self, session_id: str) -> Dict[str, Any]:
        """创建 Session"""
        resp = requests.post(
            f'{self.base_url}/api/sessions',
            headers=self._headers(),
            json={'session_id': session_id}
        )
        resp.raise_for_status()
        return resp.json()

    def get_session(self, session_id: str) -> Dict[str, Any]:
        """获取 Session"""
        resp = requests.get(
            f'{self.base_url}/api/sessions/{session_id}',
            headers=self._headers()
        )
        resp.raise_for_status()
        return resp.json()

    def list_sessions(self) -> Dict[str, Any]:
        """列出所有 Session"""
        resp = requests.get(
            f'{self.base_url}/api/sessions',
            headers=self._headers()
        )
        resp.raise_for_status()
        return resp.json()

    def delete_session(self, session_id: str) -> None:
        """删除 Session"""
        resp = requests.delete(
            f'{self.base_url}/api/sessions/{session_id}',
            headers=self._headers()
        )
        resp.raise_for_status()

    def start_kernel(self, session_id: str) -> Dict[str, Any]:
        """启动 Kernel"""
        resp = requests.post(
            f'{self.base_url}/api/sessions/{session_id}/start',
            headers=self._headers()
        )
        resp.raise_for_status()
        return resp.json()

    def execute(self, session_id: str, code: str, timeout: int = 60) -> Dict[str, Any]:
        """执行代码"""
        resp = requests.post(
            f'{self.base_url}/api/sessions/{session_id}/execute',
            headers=self._headers(),
            json={'code': code, 'timeout': f'{timeout}s'}
        )
        resp.raise_for_status()
        return resp.json()

    def interrupt(self, session_id: str) -> None:
        """中断执行"""
        resp = requests.post(
            f'{self.base_url}/api/sessions/{session_id}/interrupt',
            headers=self._headers()
        )
        resp.raise_for_status()

    def restart(self, session_id: str) -> Dict[str, Any]:
        """重启 Kernel"""
        resp = requests.post(
            f'{self.base_url}/api/sessions/{session_id}/restart',
            headers=self._headers()
        )
        resp.raise_for_status()
        return resp.json()

    def stop_kernel(self, session_id: str) -> None:
        """停止 Kernel"""
        resp = requests.post(
            f'{self.base_url}/api/sessions/{session_id}/stop',
            headers=self._headers()
        )
        resp.raise_for_status()

    def get_kernel_info(self, session_id: str) -> Dict[str, Any]:
        """获取 Kernel 信息"""
        resp = requests.get(
            f'{self.base_url}/api/sessions/{session_id}/kernel',
            headers=self._headers()
        )
        resp.raise_for_status()
        return resp.json()

    def get_state(self, session_id: str) -> Dict[str, Any]:
        """获取状态"""
        resp = requests.get(
            f'{self.base_url}/api/sessions/{session_id}/state',
            headers=self._headers()
        )
        resp.raise_for_status()
        return resp.json()

    def restore_state(self, session_id: str, namespace: dict, execution_count: int) -> None:
        """恢复状态"""
        resp = requests.post(
            f'{self.base_url}/api/sessions/{session_id}/state',
            headers=self._headers(),
            json={
                'namespace': namespace,
                'execution_count': execution_count
            }
        )
        resp.raise_for_status()


if __name__ == '__main__':
    client = KernelHubClient(
        base_url='http://localhost:8081',
        auth_token='your-jwt-token'
    )

    client.create_session('session_123')

    client.start_kernel('session_123')

    result = client.execute('session_123', 'x = 10')
    print(f"Result: {result}")

    result = client.execute('session_123', 'y = 20')
    print(f"Result: {result}")

    result = client.execute('session_123', 'x + y')
    print(f"Result: {result}")

    state = client.get_state('session_123')
    print(f"State: {state}")

    client.stop_kernel('session_123')
    client.delete_session('session_123')
