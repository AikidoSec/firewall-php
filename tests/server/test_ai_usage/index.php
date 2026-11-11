<?php
function collectAIBytes($handle, $bytes) {
    $GLOBALS['aiCallbackBody'] .= $bytes;
    return strlen($bytes);
}

class AIWriteCallback {
    public function __invoke($handle, $bytes) {
        return collectAIBytes($handle, $bytes);
    }

    public function write($handle, $bytes) {
        return collectAIBytes($handle, $bytes);
    }
}

function captureCallableResponses($origin) {
    $bodies = [];
    foreach (['collectAIBytes', new AIWriteCallback(), [new AIWriteCallback(), 'write']] as $callback) {
        $GLOBALS['aiCallbackBody'] = '';
        $handle = curl_init($origin.'/v1/chat/completions');
        curl_setopt_array($handle, [CURLOPT_POST => true, CURLOPT_WRITEFUNCTION => $callback]);
        if (curl_exec($handle) !== true) {
            throw new RuntimeException('Callable transfer failed');
        }
        $bodies[] = hash('sha256', $GLOBALS['aiCallbackBody']);
    }
    return $bodies;
}

function captureSingleResponses($origin, $paths, $style) {
    $bodies = [];
    foreach ($paths as $path) {
        $handle = curl_init($origin.$path);
        curl_setopt($handle, CURLOPT_POST, true);
        $body = '';
        if ($style === 'return') {
            curl_setopt($handle, CURLOPT_RETURNTRANSFER, true);
        } else {
            $sink = fopen('php://temp', 'w+');
            curl_setopt_array($handle, [
                CURLOPT_RETURNTRANSFER => false,
                CURLOPT_WRITEFUNCTION => function ($handle, $bytes) use ($sink) {
                    return fwrite($sink, $bytes);
                },
            ]);
        }
        $result = curl_exec($handle);
        if ($style === 'return') {
            $body = $result;
        } else {
            rewind($sink);
            $body = stream_get_contents($sink);
            fclose($sink);
        }
        if ($result !== ($style === 'return' ? $body : true)) {
            throw new RuntimeException('curl_exec result changed');
        }
        $bodies[$path] = hash('sha256', $body);
    }
    return $bodies;
}

function captureMultiResponses($origin, $paths, $style) {
    $multi = curl_multi_init();
    $handles = [];
    $chunks = [];
    foreach ($paths as $path) {
        $handle = curl_init($origin.$path);
        curl_setopt($handle, CURLOPT_CUSTOMREQUEST, 'POST');
        if ($style === 'replacement') {
            curl_setopt($handle, CURLOPT_URL, $origin.$path.'?chunks');
        }
        $chunks[$path] = '';
        if ($style === 'return') {
            curl_setopt($handle, CURLOPT_RETURNTRANSFER, true);
        } elseif ($style === 'callback') {
            curl_setopt($handle, CURLOPT_WRITEFUNCTION, function ($handle, $bytes) use (&$chunks, $path) {
                $chunks[$path] .= $bytes;
                return strlen($bytes);
            });
        } elseif ($style === 'replacement') {
            curl_setopt($handle, CURLOPT_WRITEFUNCTION, function ($handle, $bytes) use (&$chunks, $path) {
                curl_setopt($handle, CURLOPT_WRITEFUNCTION, function ($handle, $bytes) use (&$chunks, $path) {
                    $chunks[$path] .= $bytes;
                    return strlen($bytes);
                });
                $chunks[$path] .= $bytes;
                return strlen($bytes);
            });
        }
        curl_multi_add_handle($multi, $handle);
        $handles[$path] = $handle;
    }
    do {
        if (curl_multi_exec($multi, $running) !== CURLM_OK) {
            throw new RuntimeException('curl_multi_exec failed');
        }
        if ($running) {
            curl_multi_select($multi, 1);
        }
    } while ($running);
    $completed = 0;
    while ($message = curl_multi_info_read($multi)) {
        if ($message['msg'] !== CURLMSG_DONE || $message['result'] !== CURLE_OK) {
            throw new RuntimeException('Completion changed');
        }
        $completed++;
    }
    if ($completed !== count($paths)) {
        throw new RuntimeException('Completion queue changed');
    }
    $bodies = [];
    foreach ($handles as $path => $handle) {
        if ($style === 'return') {
            $body = curl_multi_getcontent($handle);
            if (curl_multi_getcontent($handle) !== $body) {
                throw new RuntimeException('Multi return buffer was consumed');
            }
        } elseif ($style === 'callback' || $style === 'replacement') {
            $body = $chunks[$path];
        }
        $bodies[$path] = hash('sha256', $body);
    }
    foreach ($handles as $handle) {
        curl_multi_remove_handle($multi, $handle);
    }
    return $bodies;
}

function checkFailedCallbacks($origin) {
    $handle = curl_init($origin.'/v1/chat/completions');
    curl_setopt_array($handle, [
        CURLOPT_POST => true,
        CURLOPT_WRITEFUNCTION => function ($handle, $bytes) {
            return 0;
        },
    ]);
    if (curl_exec($handle) !== false || curl_errno($handle) !== CURLE_WRITE_ERROR) {
        throw new RuntimeException('Callback abort changed');
    }
    $exception = new RuntimeException('Application callback error');
    curl_setopt($handle, CURLOPT_WRITEFUNCTION, function ($handle, $bytes) use ($exception) {
        throw $exception;
    });
    try {
        curl_exec($handle);
        throw new RuntimeException('Callback exception was swallowed');
    } catch (RuntimeException $caught) {
        if ($caught !== $exception) {
            throw $caught;
        }
    }
}

function captureSkippedAndCompressedResponses($origin) {
    $bodies = [];
    $handle = curl_init();
    curl_setopt_array($handle, [CURLOPT_POST => true, CURLOPT_RETURNTRANSFER => true]);
    foreach (['error', 'invalid', 'oversized', 'deep'] as $query) {
        curl_setopt($handle, CURLOPT_URL, $origin.'/v1/chat/completions?'.$query);
        $bodies['skip'][$query] = hash('sha256', curl_exec($handle));
    }
    foreach ([
        '/other',
        '/publishers/other/models/claude:rawPredict',
        '/models/foo/bar:generateContent',
        '/model/foo/bar/converse',
    ] as $path) {
        curl_setopt($handle, CURLOPT_URL, $origin.$path);
        $bodies['skip'][$path] = hash('sha256', curl_exec($handle));
    }
    curl_setopt_array($handle, [CURLOPT_URL => $origin.'/v1/chat/completions?gzip', CURLOPT_ENCODING => '']);
    $bodies['gzip'] = hash('sha256', curl_exec($handle));
    curl_setopt_array($handle, [CURLOPT_URL => $origin.'/redirect', CURLOPT_FOLLOWLOCATION => true, CURLOPT_POSTFIELDS => '']);
    $bodies['redirect'] = hash('sha256', curl_exec($handle));
    return $bodies;
}

if (!isset($_GET['provider'])) {
    header('Content-Type: application/json');
    echo json_encode(['phpVersion' => PHP_VERSION]);
    return;
}

$origin = $_GET['provider'];
$paths = [
    '/v1/chat/completions',
    '/v1/responses',
    '/v1/messages',
    '/v1beta/models/gemini:generateContent',
    '/model/anthropic.claude/converse',
    '/model/anthropic.claude/invoke',
];
$bodies = [];
foreach (['return', 'callback'] as $style) {
    $bodies[$style] = captureSingleResponses($origin, $paths, $style);
}
foreach (['return', 'callback', 'replacement'] as $style) {
    $bodies['multi-'.$style] = captureMultiResponses($origin, $paths, $style);
}
$bodies['callables'] = captureCallableResponses($origin);
checkFailedCallbacks($origin);
$bodies = array_merge($bodies, captureSkippedAndCompressedResponses($origin));

header('Content-Type: application/json');
echo json_encode($bodies);
