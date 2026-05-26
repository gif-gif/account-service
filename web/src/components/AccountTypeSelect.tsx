import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "./ui/select";
import { accountTypes, type AccountType } from "../store/accounts";

const ALL_ACCOUNT_TYPES_VALUE = "__all_account_types__";

type AccountTypeSelectProps = {
  allLabel?: string;
  ariaLabel: string;
  defaultValue?: AccountType | "";
  includeAll?: boolean;
  name?: string;
  onValueChange?: (value: string) => void;
  value?: string;
};

export function AccountTypeSelect({ allLabel, ariaLabel, defaultValue, includeAll = false, name, onValueChange, value }: AccountTypeSelectProps) {
  const selectedValue = value !== undefined ? normalizeValue(value, includeAll) : undefined;
  const initialValue = value === undefined ? normalizeValue(defaultValue ?? accountTypes[0], includeAll) : undefined;

  return (
    <Select
      defaultValue={initialValue}
      name={name}
      value={selectedValue}
      onValueChange={(nextValue) => onValueChange?.(nextValue === ALL_ACCOUNT_TYPES_VALUE ? "" : nextValue)}
    >
      <SelectTrigger aria-label={ariaLabel}>
        <SelectValue />
      </SelectTrigger>
      <SelectContent>
        {includeAll ? <SelectItem value={ALL_ACCOUNT_TYPES_VALUE}>{allLabel ?? "All"}</SelectItem> : null}
        {accountTypes.map((accountType) => (
          <SelectItem key={accountType} value={accountType}>
            {accountType}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}

function normalizeValue(value: string, includeAll: boolean) {
  if (includeAll && value === "") {
    return ALL_ACCOUNT_TYPES_VALUE;
  }
  return value || accountTypes[0];
}
