import gzip
import json
from contextlib import contextmanager
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from threading import Thread
from urllib.parse import urlsplit


RESPONSES = {
    '/v1/chat/completions': {
        'id': 'chatcmpl-test', 'object': 'chat.completion', 'model': 'gpt-4o',
        'choices': [{'message': {'role': 'assistant', 'content': 'Hello'}}],
        'usage': {'prompt_tokens': 10, 'completion_tokens': 5,
                  'prompt_tokens_details': {'cached_tokens': 8}},
    },
    '/v1/responses': {
        'id': 'resp-test', 'model': 'gpt-4o', 'output': [],
        'usage': {'input_tokens': 10, 'output_tokens': 5},
    },
    '/v1/messages': {
        'id': 'msg-test', 'type': 'message', 'model': 'claude',
        'content': [{'type': 'text', 'text': 'Hello'}],
        'usage': {'input_tokens': 10, 'output_tokens': 5, 'cache_creation_input_tokens': 8},
    },
    '/v1beta/models/gemini:generateContent': {
        'modelVersion': 'gemini',
        'candidates': [{'content': {'parts': [{'text': 'Hello'}]}}],
        'usageMetadata': {'promptTokenCount': 10, 'candidatesTokenCount': 3, 'thoughtsTokenCount': 2},
    },
    '/model/anthropic.claude/converse': {
        'output': {'message': {'role': 'assistant', 'content': [{'text': 'Hello'}]}},
        'stopReason': 'end_turn',
        'usage': {'inputTokens': 10, 'outputTokens': 5, 'totalTokens': 15, 'cacheReadInputTokens': 8},
    },
    '/model/anthropic.claude/invoke': {
        'type': 'message', 'content': [{'type': 'text', 'text': 'Hello'}],
        'usage': {'input_tokens': 10, 'output_tokens': 5},
    },
}


def response_bytes(path, query=''):
    if query == 'invalid':
        return b'{invalid'
    if query == 'oversized':
        return json.dumps({'model': 'gpt-4o', 'text': 'x' * (2 * 1024 * 1024)}).encode()
    if query == 'chunks':
        body = dict(RESPONSES[path], padding='x' * (64 * 1024))
        return json.dumps(body).encode()
    if query == 'deep':
        return b'{"model":"gpt-4o","data":' + b'[' * 1000 + b'0' + b']' * 1000 + b'}'
    return json.dumps(RESPONSES.get(path, RESPONSES['/v1/chat/completions'])).encode()


class ProviderHandler(BaseHTTPRequestHandler):
    def do_POST(self):
        self.rfile.read(int(self.headers.get('Content-Length', 0)))
        url = urlsplit(self.path)
        if url.path == '/redirect':
            self.send_response(307)
            self.send_header('Location', '/v1/chat/completions')
            self.send_header('Content-Length', '0')
            self.end_headers()
            return
        body = response_bytes(url.path, url.query)
        self.send_response(401 if url.query == 'error' else 200)
        self.send_header('Content-Type', 'application/json')
        if url.query == 'gzip':
            body = gzip.compress(body)
            self.send_header('Content-Encoding', 'gzip')
        self.send_header('Content-Length', str(len(body)))
        self.end_headers()
        for offset in range(0, len(body), 64 * 1024):
            self.wfile.write(body[offset:offset + 64 * 1024])
            self.wfile.flush()

    def log_message(self, *args):
        pass


@contextmanager
def serve_provider():
    server = ThreadingHTTPServer(('127.0.0.1', 0), ProviderHandler)
    thread = Thread(target=server.serve_forever)
    thread.start()
    try:
        yield f'http://127.0.0.1:{server.server_port}'
    finally:
        server.shutdown()
        thread.join()
        server.server_close()
