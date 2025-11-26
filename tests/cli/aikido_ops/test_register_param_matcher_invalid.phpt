--TEST--
Test \aikido\register_param_matcher with invalid parameters

--ENV--
AIKIDO_LOG_LEVEL=INFO
AIKIDO_TOKEN=AIK_DUMMY

--FILE--
<?php

$invalidPatterns = [
    "no_braces"      => "digits-alpha",      // missing { }
    "unclosed_brace" => "{digits",           // only opening brace
    "with_slash"     => "aikido/{digits}",   // contains a slash, which is not allowed
    "invalid_regex"  => ".*[a-z]*-{abcd}",    // does not contain {digits} or {alpha}
];

foreach ($invalidPatterns as $name => $pattern) {
    $result = \aikido\register_param_matcher($name, $pattern);
    var_dump($result);
}

?>

--EXPECT--
[AIKIDO][INFO] Token changed to "AIK_RUNTIME_***UMMY"

003+ Error compiling param matcher no_braces -> regex "digits-alpha": pattern should contain { or }
bool(false)
006+ Error compiling param matcher unclosed_brace -> regex "{digits": pattern should contain { or }
bool(false)
009+ Error compiling param matcher with_slash -> regex "aikido/{digits}": pattern should not contain slashes
bool(false)
