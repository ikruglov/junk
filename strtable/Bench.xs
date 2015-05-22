#define PERL_NO_GET_CONTEXT
#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
#include "ppport.h"

#include "strtable.h"

MODULE = Bench        PACKAGE = Bench

void
benchmark_strtable_leveldb_hash(data, size)
    AV *data;
    UV size;
  PREINIT:
    SV *sv;
    STRLEN len;
    const char *str;
    SSize_t i, avlen;
    strtable_t tbl;
  CODE:
    avlen = av_len(data) + 1;
    strtable_init(&tbl, (uint32_t) size);

    for (i = 0; i < avlen; ++i) {
        sv = *av_fetch(data, i, 0);
        str = SvPV(sv, len);
        strtable_insert_leveldb_hash(&tbl, str, (uint32_t) len, (uint32_t) i);
    }

    //printf("strtable_leveldb_hash_%d: table containt %d keys\n", (int) size, tbl.keys);
    strtable_clear(&tbl);

void
benchmark_strtable_perl_hash(data, size)
    AV *data;
    UV size;
  PREINIT:
    SV *sv;
    STRLEN len;
    uint32_t hash;
    const char *str;
    SSize_t i, avlen;
    strtable_t tbl;
  CODE:
    avlen = av_len(data) + 1;
    strtable_init(&tbl, (uint32_t) size);

    for (i = 0; i < avlen; ++i) {
        sv = *av_fetch(data, i, 0);
        str = SvPV(sv, len);
        PERL_HASH(hash, str, len);
        strtable_insert(&tbl, hash, str, (uint32_t) len, (uint32_t) i);
    }

    //printf("strtable_perl_hash_%d: table containt %d keys\n", (int) size, tbl.keys);
    strtable_clear(&tbl);

void 
benchmark_hv(data)
    AV *data;
  PREINIT:
    SV *sv;
    HV *hv;
    STRLEN len;
    const char *str;
    SSize_t i, avlen;
  CODE:
    hv = newHV();
    avlen = av_len(data) + 1;

    for (i = 0; i < avlen; ++i) {
        sv = *av_fetch(data, i, 0);
        str = SvPV(sv, len);
        hv_fetch(hv, str, len, 1);
    }

    //printf("benchmark_hv: table containt %d keys\n", (int) av_len((AV*) hv));
    hv_undef(hv);
