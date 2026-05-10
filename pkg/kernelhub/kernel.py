#!/usr/bin/env python3

import sys
import json
import signal
import resource
from typing import Dict


class Kernel:
    def __init__(self):
        self.user_ns = {}
        self.execution_count = 0
        self._memory_limit = 512 * 1024 * 1024

    def execute(self, code: str, timeout: int = 60) -> Dict:
        self.execution_count += 1

        stdout_buffer = []
        stderr_buffer = []

        old_stdout = sys.stdout
        old_stderr = sys.stderr

        try:
            sys.stdout = BufferedWriter(stdout_buffer)
            sys.stderr = BufferedWriter(stderr_buffer)

            compiled = compile(code, '<kernel>', 'exec')
            exec(compiled, self.user_ns)

            status = "ok"

        except MemoryError:
            status = "error"
            stderr_buffer.append("Memory limit exceeded\n")

        except Exception as e:
            status = "error"
            stderr_buffer.append(f"{type(e).__name__}: {e}\n")

        finally:
            sys.stdout = old_stdout
            sys.stderr = old_stderr

        return {
            "type": "execute_reply",
            "content": {
                "status": status,
                "count": self.execution_count,
                "stdout": "".join(stdout_buffer),
                "stderr": "".join(stderr_buffer),
            }
        }

    def get_namespace(self) -> Dict:
        safe_ns = {}
        for k, v in self.user_ns.items():
            if not k.startswith('_'):
                try:
                    json.dumps(v)
                    safe_ns[k] = v
                except (TypeError, ValueError):
                    safe_ns[k] = f"<{type(v).__name__}>"
        return safe_ns

    def restore_namespace(self, ns: Dict) -> None:
        self.user_ns.update(ns)


class BufferedWriter:
    def __init__(self, buffer):
        self.buffer = buffer

    def write(self, text):
        self.buffer.append(text)

    def flush(self):
        pass


def main():
    import argparse
    parser = argparse.ArgumentParser()
    parser.add_argument('--session-id', default='')
    parser.add_argument('--memory-limit', type=int, default=512*1024*1024)
    args = parser.parse_args()

    kernel = Kernel()
    kernel._memory_limit = args.memory_limit

    soft, hard = resource.getrlimit(resource.RLIMIT_AS)
    resource.setrlimit(resource.RLIMIT_AS, (args.memory_limit, hard))

    signal.signal(signal.SIGINT, lambda s, f: None)

    print(json.dumps({"type": "ready"}), flush=True)

    for line in sys.stdin:
        line = line.strip()
        if not line:
            continue

        try:
            msg = json.loads(line)
        except:
            continue

        msg_type = msg.get('type')
        msg_id = msg.get('msg_id', '')
        content = msg.get('content', {})

        if msg_type == 'execute':
            code = content.get('code', '')
            timeout = content.get('timeout', 60)
            result = kernel.execute(code, timeout)
            result['msg_id'] = msg_id
            print(json.dumps(result), flush=True)

        elif msg_type == 'get_namespace':
            print(json.dumps({
                "type": "namespace",
                "msg_id": msg_id,
                "content": kernel.get_namespace()
            }), flush=True)

        elif msg_type == 'restore_namespace':
            kernel.restore_namespace(content)
            print(json.dumps({
                "type": "restore_reply",
                "msg_id": msg_id,
                "content": {"status": "ok"}
            }), flush=True)

        elif msg_type == 'interrupt':
            print(json.dumps({
                "type": "interrupt_reply",
                "msg_id": msg_id,
                "content": {"status": "ok"}
            }), flush=True)


if __name__ == '__main__':
    main()
