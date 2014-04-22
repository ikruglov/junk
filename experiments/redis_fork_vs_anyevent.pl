#!/usr/bin/env perl

use strict;
use warnings;

use Redis;
use AnyEvent::Redis;
use Benchmark qw/:all/;
use POSIX ":sys_wait_h";
use Data::Dumper;

my $data = 'A' x (10 * 1024 * 1024); # 10MB
my @hosts = ( 'localhost:6379', 'localhost:6381', 'localhost:6383' );

cmpthese(10, {
    fork => sub {
        foreach my $host (@hosts) {
            my $pid = fork();
            next if $pid; # exit in parent

            my $redis = Redis->new(server => $host);
            $redis->set("_test_$_", $data) foreach (1..10);

            undef $redis;
            POSIX::exit 0; # exit in child
        }

        waitpid(-1, 0) foreach (@hosts);
    },

    anyevent => sub {
        my @redises;
        foreach my $server (@hosts) {
            my ($host, $port) = split(/:/, $server);
            my $redis = AnyEvent::Redis->new(
                host => $host,
                port => $port,
                on_error => sub { warn @_ },
                on_cleanup => sub { warn "Connection closed: @_" },
            );

            $redis->set("_test_$_", $data) foreach (1..10);
            $redis->all_cv()->recv();
            push @redises, $redis;
        }
    },
});

exit 0;
