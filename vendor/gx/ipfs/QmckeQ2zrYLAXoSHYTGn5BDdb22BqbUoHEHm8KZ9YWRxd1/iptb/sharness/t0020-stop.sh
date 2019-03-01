#!/bin/sh

test_description="iptb stop tests"

. lib/test-lib.sh

export IPTB_ROOT=.
ln -s ../plugins $IPTB_ROOT/plugins

test_expect_success "iptb auto works" '
	../bin/iptb auto -count 3 -type localipfs
'

test_expect_success "iptb start works" '
	../bin/iptb start --wait -- --debug
'

test_expect_success "iptb stop works" '
	../bin/iptb stop && sleep 1
'

for i in {0..2}; do
	test_expect_success "daemon '$i' was shut down gracefully" '
		cat testbeds/default/'$i'/daemon.stderr | tail -1 | grep "Gracefully shut down daemon"
	'
done

test_done
