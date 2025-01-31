#!/bin/bash

#!/bin/bash

set -u

source './build.cfg' || exit 1

echo 'removing executables'
find bin -type f -name "$executable" -exec rm -v '{}' \;
