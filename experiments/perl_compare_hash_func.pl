#!/usr/bin/env perl

use strict;
use warnings;

use Digest::MD5 qw/md5_hex/;
use Digest::SHA qw/sha1_hex sha256_hex sha512_hex/;
use Benchmark qw/timethese/;

$/ = undef;
my $data = <STDIN>;
printf "length: %d\n", length($data);

timethese(-10, {
    md5    => sub { md5_hex($data) },
    sha1   => sub { sha1_hex($data) },
    sha256 => sub {sha256_hex($data) },
    sha512 => sub { sha256_hex($data) },
});
