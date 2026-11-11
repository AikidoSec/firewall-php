import hashlib
from urllib.parse import urlencode

from ai_provider_server import response_bytes, serve_provider
from testlib import load_test_args, php_server_get


def run_test():
    with serve_provider() as origin:
        response = php_server_get('/?' + urlencode({'provider': origin}))
        assert response.status_code == 200, response.text
        assert response.json() == {
            'ok': True,
            'bodyHash': hashlib.sha256(response_bytes('/other')).hexdigest(),
            'completions': 4,
        }


if __name__ == '__main__':
    load_test_args()
    run_test()
