#!/usr/local/bin/booking-perl

use 5.14.2;
use strict;
use warnings;

use Time::HiRes;
use Sereal::Encoder;
use Sereal::Decoder;
use Getopt::Long qw(GetOptions);

my (
    $dump_to,
    $dedupe_strings,
);

BEGIN {
    GetOptions(
        'dump_to|dump-to=s'             => \$dump_to,
        'dedupe_strings|dedupe-strings' => \$dedupe_strings,
    );
}

my $decoder = Sereal::Decoder->new();
my $encoder = Sereal::Encoder->new({
    dedupe_strings => $dedupe_strings
});

$/ = undef;
my @data;

say "read and deserialize data";
foreach my $file (glob($ARGV[0] . "/*.srl")) {
    open(my $fh, '<', $file) or die $!;
    my $content = <$fh>;
    close($fh);

    my $decoded = $decoder->decode($content);
    my $ref = ref $decoded;

    if ($ref eq 'ARRAY') {
        push @data, @$decoded;
    } else {
        push @data, $decoded;
    }
}

say "serialize data";
my $start = Time::HiRes::time();
my $result = $encoder->encode(\@data, { dedupe_strings => 1 });
printf("Call tool %.2f seconds\n", Time::HiRes::time() - $start);

if ($dump_to) {
    say "Dumping result to $dump_to\n";
    open(my $fh, '>', $dump_to) or die $!;
    print $fh $result;
    close $fh;
}

exit 0;
