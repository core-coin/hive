# Removes all empty keys and values in input.
def remove_empty:
  . | walk(
    if type == "object" then
      with_entries(
        select(
          .value != null and
          .value != "" and
          .value != [] and
          .key != null and
          .key != ""
        )
      )
    else .
    end
  )
;

# Converts decimal string to number.
def to_int:
  if . == null then . else .|tonumber end
;

# Converts "1" / "0" to boolean.
def to_bool:
  if . == null then . else
    if . == "1" then true else false end
  end
;

# Replace config in input.
. + {
  "config": {
    "cryptore": (if env.HIVE_CLIQUE_PERIOD then null else {} end),
    "networkId": (if env.HIVE_NETWORK_ID then env.HIVE_NETWORK_ID|to_int else 1337 end),
  }|remove_empty
}
