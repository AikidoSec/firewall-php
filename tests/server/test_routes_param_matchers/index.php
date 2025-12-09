<?php

// Register custom param matchers on each request. The backend will ignore
// duplicates, so this is safe to call per request.
\aikido\register_param_matcher("tenant", "aikido-{digits}");
\aikido\register_param_matcher("slug", "aikido-{alpha}-{digits}-{alpha}");

echo "OK";

?>

