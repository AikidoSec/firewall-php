import hashlib
from urllib.parse import urlencode

from testlib import (
    load_test_args, mock_server_get_events,
    php_server_get, wait_until,
)
from ai_provider_server import RESPONSES, response_bytes, serve_provider


def run_test():
    version = tuple(int(part) for part in php_server_get('/').json()['phpVersion'].split('.')[:2])
    if version < (8, 0):
        return

    # Check real transport capture, IPC, and aggregation in the emitted heartbeat.
    with serve_provider() as origin:
        response = php_server_get('/?' + urlencode({'provider': origin}))
        assert response.status_code == 200, response.text
        actual_bodies = response.json()
        expected = {path: hashlib.sha256(response_bytes(path)).hexdigest() for path in RESPONSES}
        for style in ['return', 'callback', 'multi-return', 'multi-callback']:
            assert actual_bodies[style] == expected, style
        assert actual_bodies['multi-replacement'] == {
            path: hashlib.sha256(response_bytes(path, 'chunks')).hexdigest() for path in RESPONSES
        }
        assert actual_bodies['skip'] == {
            **{query: hashlib.sha256(response_bytes('/v1/chat/completions', query)).hexdigest()
               for query in ['error', 'invalid', 'oversized', 'deep']},
            **{path: hashlib.sha256(response_bytes(path)).hexdigest() for path in [
                '/other',
                '/publishers/other/models/claude:rawPredict',
                '/models/foo/bar:generateContent',
                '/model/foo/bar/converse',
            ]},
        }
        assert actual_bodies['gzip'] == expected['/v1/chat/completions']
        assert actual_bodies['redirect'] == expected['/v1/chat/completions']
        assert actual_bodies['callables'] == [expected['/v1/chat/completions']] * 3

        heartbeats = []

        def received_heartbeat():
            heartbeats[:] = [event for event in mock_server_get_events() if event.get('type') == 'heartbeat']
            return bool(heartbeats)

        assert wait_until(received_heartbeat, 90), 'No heartbeat received'
        actual = sorted(heartbeats[0]['ai'], key=lambda row: row['model'])
        expected_ai = [
            {'provider': origin, 'model': 'gpt-4o', 'calls': 15,
             'tokens': {'input': 150, 'output': 75, 'total': 225, 'cacheRead': 80, 'cacheWrite': 0}},
            {'provider': origin, 'model': 'claude', 'calls': 5,
             'tokens': {'input': 50, 'output': 25, 'total': 75, 'cacheRead': 0, 'cacheWrite': 40}},
            {'provider': origin, 'model': 'gemini', 'calls': 5,
             'tokens': {'input': 50, 'output': 25, 'total': 75, 'cacheRead': 0, 'cacheWrite': 0}},
            {'provider': origin, 'model': 'anthropic.claude', 'calls': 10,
             'tokens': {'input': 100, 'output': 50, 'total': 150, 'cacheRead': 40, 'cacheWrite': 0}},
        ]
        if version < (8, 2):
            expected_ai = [
                {'provider': origin, 'model': 'gpt-4o', 'calls': 6,
                 'tokens': {'input': 60, 'output': 30, 'total': 90, 'cacheRead': 32, 'cacheWrite': 0}},
                {'provider': origin, 'model': 'claude', 'calls': 2,
                 'tokens': {'input': 20, 'output': 10, 'total': 30, 'cacheRead': 0, 'cacheWrite': 16}},
                {'provider': origin, 'model': 'gemini', 'calls': 2,
                 'tokens': {'input': 20, 'output': 10, 'total': 30, 'cacheRead': 0, 'cacheWrite': 0}},
                {'provider': origin, 'model': 'anthropic.claude', 'calls': 4,
                 'tokens': {'input': 40, 'output': 20, 'total': 60, 'cacheRead': 16, 'cacheWrite': 0}},
            ]
        assert actual == sorted(expected_ai, key=lambda row: row['model']), actual
        assert 'aiTools' not in heartbeats[0]


if __name__ == '__main__':
    load_test_args()
    run_test()
