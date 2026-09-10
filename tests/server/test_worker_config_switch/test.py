from pathlib import Path
from testlib import *


def run_test():
    workers = set()
    root = Path(__file__).resolve().parent
    for phase, (site, blocked) in enumerate([('a', False), ('b', True), ('c', False), ('none', False), ('a', False)]):
        document_root = root if site == 'none' else root / site
        response = php_server_get('/?path=../test_worker_config_switch/index.php',
                                  headers={'X-Test-Root': str(document_root)})
        assert_response_code_is(response, 200)
        result = response.json()
        if 'skip' in result:
            print(f"Skipping: {result['skip']}")
            return
        assert result['blocked'] == blocked, f'Site {site}: {result}'
        print(f"Site {site}: blocked={blocked}, worker={result['worker']}")
        if phase < 3:
            workers.add(result['worker'])
            token = mock_server_get_token()
            assert token == f'worker-site-{site}', f'Site {site}: reporting token was {token}'
    # Three configured sites and two workers guarantee a cross-site switch.
    assert len(workers) < 3, 'Each configured site was handled by a different worker'


if __name__ == '__main__':
    load_test_args()
    run_test()
