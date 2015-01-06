Generate a trace

```php
ini_set( 'xdebug.collect_params', 3 );
xdebug_start_trace("/tmp/trace.out");

... do stuff

xdebug_stop_trace();
```

Download your trace

use the -f flag to examine your trace

go to http://127.0.0.1:8888/
