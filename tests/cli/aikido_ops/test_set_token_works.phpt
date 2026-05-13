--TEST--	
Test \aikido\set_token with a valid token	

--ENV--	
AIKIDO_LOG_LEVEL=INFO	

--FILE--	
<?php	

\aikido\set_token("your token here")	

?>	

--EXPECTREGEX--	
\[AIKIDO\]\[INFO\]\[tid:\d+\] Token changed to "AIK_RUNTIME_\*\*\*here"
