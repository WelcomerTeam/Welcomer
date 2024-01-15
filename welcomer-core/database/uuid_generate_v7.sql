create or replace function uuid_generate_v7()
returns uuid
as $$
begin
  return encode(overlay(
    set_bit(set_bit(uuid_send(gen_random_uuid()), 53, 1), 52, 1)
    placing substring(int8send((extract(epoch from clock_timestamp()) * 1000)::bigint) from 3)
    from 1 for 6), 'hex')::uuid;
end
$$
language plpgsql
volatile;