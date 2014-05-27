#!/usr/bin/env perl

use strict;
use warnings;

use JSON::XS;
use Data::Dumper;
use Sereal::Encoder qw/encode_sereal/;
use Sereal::Decoder qw/decode_sereal/;

$/ = undef;
my $data_in = <STDIN>;

my $decoded_data = decode_sereal($data_in);
print encode_json($decoded_data);
#print encode_sereal($decoded_data, {});

exit 0;
