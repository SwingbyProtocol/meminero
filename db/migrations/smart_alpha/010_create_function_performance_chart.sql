create or replace function smart_alpha.performance_at_ts(pool text, ts bigint)
    returns table
            (
                senior_without_sa double precision,
                senior_with_sa    double precision,
                junior_without_sa double precision,
                junior_with_sa    double precision
            )
    language plpgsql
as
$$
declare
    senior_without_sa  double precision;
    senior_with_sa     double precision;
    junior_without_sa  double precision;
    junior_with_sa     double precision;
    token_price        double precision;
    token_address      text;
    token_decimals     integer;
    quote_asset_symbol text;
begin
    select into token_address,quote_asset_symbol,token_decimals p.pool_token_address,
                                                                p.oracle_asset_symbol,
                                                                p.pool_token_decimals
    from smart_alpha.pools p
    where pool_address = pool;
    select into token_price public.token_price_at_ts(token_address, quote_asset_symbol, ts);

    select into senior_without_sa (((select senior_liquidity
                                     from smart_alpha.pool_epoch_info
                                     where block_timestamp <= ts
                                     order by block_timestamp desc
                                     limit 1)::numeric(78, 18) / pow(10, token_decimals)) * token_price);

    select into senior_with_sa (((select estimated_senior_liquidity
                                  from smart_alpha.pool_state
                                  where block_timestamp <= ts
                                  order by block_timestamp desc
                                  limit 1)::numeric(78, 18) / pow(10, token_decimals)) * token_price);
    select into junior_without_sa (((select junior_liquidity
                                     from smart_alpha.pool_epoch_info
                                     where block_timestamp <= ts
                                     order by block_timestamp desc
                                     limit 1)::numeric(78, 18) / pow(10, token_decimals)) * token_price);
    select into junior_with_sa (((select estimated_junior_liquidity
                                  from smart_alpha.pool_state
                                  where block_timestamp <= ts
                                  order by block_timestamp desc
                                  limit 1)::numeric(78, 18) / pow(10, token_decimals)) * token_price);

    return query select senior_without_sa,
                        senior_with_sa,
                        junior_without_sa,
                        junior_with_sa;
end
$$;


create or replace function smart_alpha.junior_token_to_usd_at_ts(token_address text, amount numeric(78), ts bigint) returns double precision
    language plpgsql
as
$$
declare
    oracle_asset_symbol text;
    pool_token_decimals integer;
    pool_token_address  text;
begin
    select into oracle_asset_symbol,pool_token_decimals,pool_token_address p.oracle_asset_symbol,
                                                                           p.pool_token_decimals,
                                                                           p.pool_token_address
    from smart_alpha.pools p
    where p.junior_token_address = token_address;

    return (
        select amount::numeric(78, 18) / pow(10, pool_token_decimals) *
               (select estimated_junior_token_price::numeric(78, 18) / pow(10, 18)
                from smart_alpha.pool_state
                where block_timestamp <= ts
                order by block_timestamp desc
                limit 1) * (select token_usd_price_at_ts(pool_token_address, ts))
    );
end;
$$;

