#!/usr/bin/env perl

use strict;
use warnings;

use Redis;
use AnyEvent;
use AnyEvent::Redis;
use AnyEvent::Hiredis;
use POSIX ":sys_wait_h";
use Benchmark qw/cmpthese/;

*STDOUT->autoflush();
*STDERR->autoflush();

my $n = $ARGV[0];
my $data = 'A' x $ARGV[1];
my @hosts = ( 'localhost:6379', 'localhost:6381', 'localhost:6383' );
print "N: $n, DATA LENGTH: " . length($data) . "\n";

cmpthese(10, {
    fork => sub {
        foreach my $host (@hosts) {
            my $pid = fork();
            next if $pid; # exit in parent

            my $redis = Redis->new(server => $host);
            $redis->set("_test_$_", $data) foreach (1..$n);

            undef $redis;
            POSIX::exit 0; # exit in child
        }

        waitpid(-1, 0) foreach (@hosts);
    },

    anyevent_redis => sub {
        my @redises;
        my $cond = AnyEvent->condvar;
        foreach my $server (@hosts) {
            my ($host, $port) = split(/:/, $server);
            my $redis = AnyEvent::Redis->new(
                host => $host,
                port => $port,
                on_error => sub { warn @_ },
                on_cleanup => sub { warn "Connection closed: @_" },
            );

            foreach (1..$n) {
                $cond->begin(sub { $cond->send() });
                $redis->set("_test_$_", $data, sub  { $cond->end() });
            }

            push @redises, $redis;
        }

        $cond->recv();
    },

    anyevent_hiredis => sub {
        my @redises;
        my $cond = AnyEvent->condvar;
        foreach my $server (@hosts) {
            my ($host, $port) = split(/:/, $server);
            my $redis = AnyEvent::Hiredis->new(
                host => $host,
                port => $port,
            );

            foreach (1..$n) {
                $cond->begin(sub { $cond->send() });
                $redis->command([ 'SET', "_test_$_", $data], sub { $cond->end() });
            }

            push @redises, $redis;
        }

        $cond->recv();
    },
});

exit 0;
