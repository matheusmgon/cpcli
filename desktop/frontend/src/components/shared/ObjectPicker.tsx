import { useEffect, useState } from "react";
import { useQueries, useQuery, useQueryClient } from "@tanstack/react-query";
import { Check, ChevronDown, Loader2, Plus, X } from "lucide-react";
import { toast } from "sonner";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  creatableTypes,
  findCategory,
  objectCategories,
  type CategoryKey,
  type ObjectCategory,
} from "@/config/objectCategories";
import { objectKinds } from "@/config/objectKinds";
import { getService, type JsonRecord } from "@/lib/wailsService";
import { useSessionStore } from "@/stores/session";

interface ObjectPickerProps {
  /** Selected object names. */
  value: string[];
  onChange: (names: string[]) => void;
  placeholder?: string;
  /** Which categories this field's objects can come from (e.g. Source/
   * Destination → ["network"], Service → ["service"]). Defaults to every
   * category — used by fields that legitimately accept anything. */
  categories?: CategoryKey[];
}

function useDebouncedValue(value: string, delayMs: number): string {
  const [debounced, setDebounced] = useState(value);
  useEffect(() => {
    const id = setTimeout(() => setDebounced(value), delayMs);
    return () => clearTimeout(id);
  }, [value, delayMs]);
  return debounced;
}

/** Pulls the first non-empty string among the given keys — used to surface
 * whichever "address" field a result actually has (hosts use `ipv4-address`,
 * networks use `subnet4`, ranges use `ipv4-address-first`, ...). */
function firstText(row: JsonRecord, keys: string[]): string {
  for (const key of keys) {
    const v = row[key];
    if (typeof v === "string" && v) return v;
    if (typeof v === "number") return String(v);
  }
  return "";
}

/** The API only accepts one `type` per `show-objects` call, so searching "a
 * category" means firing one call per type and merging — same technique
 * `CountObjects` callers use for the tab counters. Dedupes by uid (falling
 * back to type+name) since a type can't legitimately appear under two
 * category types at once, but defensive either way. */
async function searchAcrossTypes(filter: string, types: string[]): Promise<JsonRecord[]> {
  const service = getService();
  const batches = await Promise.all(types.map((type) => service.searchObjects(filter, type)));
  const seen = new Set<string>();
  const merged: JsonRecord[] = [];
  for (const batch of batches) {
    for (const obj of batch) {
      const name = typeof obj.name === "string" ? obj.name : "";
      if (!name) continue;
      const key = typeof obj.uid === "string" ? obj.uid : `${String(obj.type)}:${name}`;
      if (seen.has(key)) continue;
      seen.add(key);
      merged.push(obj);
    }
  }
  merged.sort((a, b) => String(a.name ?? "").localeCompare(String(b.name ?? "")));
  return merged.slice(0, 50);
}

/** Search-and-select combobox for Check Point objects (hosts, networks,
 * groups, services, ...) — replaces free-text "type the exact name" inputs
 * on Source/Destination/Service-style fields with the same live-search flow
 * SmartConsole's own object picker uses.
 *
 * v2 (previous version only did flat text search across all types): adds a
 * category bar with live counts (SmartConsole-style "Network Objects 17"),
 * shows IP/comment under each result the same way SmartConsole's dropdown
 * does, and lets the user create a missing object inline instead of leaving
 * the rule dialog to do it on the Objects page first. */
export function ObjectPicker({ value, onChange, placeholder = "Search objects...", categories }: ObjectPickerProps) {
  const queryClient = useQueryClient();
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const debouncedQuery = useDebouncedValue(query, 250);
  const [activeKey, setActiveKey] = useState<string>("all");
  const [creating, setCreating] = useState(false);
  const [createType, setCreateType] = useState("");
  const [createValues, setCreateValues] = useState<Record<string, string>>({});
  const [createSubmitting, setCreateSubmitting] = useState(false);

  const availableCategories: ObjectCategory[] = (categories && categories.length > 0
    ? categories.map(findCategory)
    : objectCategories
  ).filter((c): c is ObjectCategory => Boolean(c));
  const showTabs = availableCategories.length > 1;

  const countQueries = useQueries({
    queries: availableCategories.map((cat) => ({
      queryKey: ["object-count", cat.key, debouncedQuery],
      queryFn: () =>
        Promise.all(cat.types.map((type) => getService().countObjects(debouncedQuery, type))).then((counts) =>
          counts.reduce((a, b) => a + b, 0),
        ),
      enabled: open,
    })),
  });
  const totalCount = countQueries.reduce((sum, q) => sum + (q.data ?? 0), 0);

  const activeCategory = showTabs && activeKey !== "all" ? findCategory(activeKey) : undefined;
  const searchTypes = activeCategory ? activeCategory.types : availableCategories.flatMap((c) => c.types);

  const { data: results = [], isFetching } = useQuery({
    queryKey: ["object-search", availableCategories.map((c) => c.key).join(","), activeKey, debouncedQuery],
    queryFn: () => searchAcrossTypes(debouncedQuery, searchTypes),
    enabled: open && debouncedQuery.trim().length >= 2 && searchTypes.length > 0,
  });

  const creatable = activeCategory
    ? creatableTypes(activeCategory)
    : Array.from(new Set(availableCategories.flatMap((c) => creatableTypes(c))));

  function toggle(name: string) {
    if (value.includes(name)) {
      onChange(value.filter((v) => v !== name));
    } else {
      onChange([...value, name]);
    }
  }

  function remove(name: string) {
    onChange(value.filter((v) => v !== name));
  }

  function openCreate() {
    const type = creatable.includes(createType) ? createType : (creatable[0] ?? "");
    setCreateType(type);
    setCreateValues({});
    setCreating(true);
  }

  async function handleCreateSubmit() {
    if (!createType) return;
    setCreateSubmitting(true);
    try {
      const created = await getService().addObject(createType, createValues);
      const name = typeof created.name === "string" && created.name ? created.name : createValues.name;
      if (name) {
        onChange(value.includes(name) ? value : [...value, name]);
        toast.success(`${objectKinds[createType]?.title ?? createType} "${name}" created`);
      }
      useSessionStore.getState().markPending(1);
      queryClient.invalidateQueries({ queryKey: ["object-search"] });
      queryClient.invalidateQueries({ queryKey: ["object-count"] });
      setCreating(false);
    } catch (error) {
      toast.error(`Failed to create object: ${error instanceof Error ? error.message : String(error)}`);
    } finally {
      setCreateSubmitting(false);
    }
  }

  return (
    <div className="flex flex-col gap-1.5">
      <Popover
        open={open}
        onOpenChange={(next) => {
          setOpen(next);
          if (!next) setCreating(false);
        }}
      >
        <PopoverTrigger asChild>
          <Button
            type="button"
            variant="outline"
            role="combobox"
            aria-expanded={open}
            className="h-9 w-full justify-between font-normal"
          >
            <span className="truncate text-muted">
              {value.length > 0 ? `${value.length} selected` : placeholder}
            </span>
            <ChevronDown className="size-4 shrink-0 opacity-60" />
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-[380px] p-0">
          {showTabs && (
            <div className="flex flex-wrap gap-1 border-b border-border p-2">
              <button
                type="button"
                onClick={() => setActiveKey("all")}
                className={`rounded-md px-2 py-1 text-xs font-medium ${
                  activeKey === "all" ? "bg-accent text-accent-foreground" : "text-muted hover:bg-surface-2"
                }`}
              >
                All ({totalCount})
              </button>
              {availableCategories.map((cat, idx) => (
                <button
                  key={cat.key}
                  type="button"
                  onClick={() => setActiveKey(cat.key)}
                  className={`rounded-md px-2 py-1 text-xs font-medium ${
                    activeKey === cat.key ? "bg-accent text-accent-foreground" : "text-muted hover:bg-surface-2"
                  }`}
                >
                  {cat.label} ({countQueries[idx]?.data ?? (countQueries[idx]?.isLoading ? "…" : 0)})
                </button>
              ))}
            </div>
          )}

          {creating ? (
            <div className="flex flex-col gap-3 p-3">
              {creatable.length > 1 && (
                <div className="flex flex-col gap-1.5">
                  <Label htmlFor="picker-create-type">Type</Label>
                  <Select
                    value={createType}
                    onValueChange={(v) => {
                      setCreateType(v);
                      setCreateValues({});
                    }}
                  >
                    <SelectTrigger id="picker-create-type">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {creatable.map((type) => (
                        <SelectItem key={type} value={type}>
                          {objectKinds[type]?.title ?? type}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              )}
              {(objectKinds[createType]?.fields ?? []).map((field) => (
                <div key={field.key} className="flex flex-col gap-1.5">
                  <Label htmlFor={`picker-create-${field.key}`}>{field.label}</Label>
                  <Input
                    id={`picker-create-${field.key}`}
                    placeholder={field.placeholder}
                    value={createValues[field.key] ?? ""}
                    onChange={(e) => setCreateValues((prev) => ({ ...prev, [field.key]: e.target.value }))}
                  />
                </div>
              ))}
              <div className="flex justify-end gap-2 pt-1">
                <Button type="button" variant="outline" size="sm" onClick={() => setCreating(false)}>
                  Cancel
                </Button>
                <Button type="button" size="sm" disabled={createSubmitting} onClick={handleCreateSubmit}>
                  Create
                </Button>
              </div>
            </div>
          ) : (
            <>
              <Command shouldFilter={false}>
                <CommandInput placeholder="Type at least 2 characters..." value={query} onValueChange={setQuery} />
                <CommandList>
                  {query.trim().length < 2 ? (
                    <CommandEmpty>Type at least 2 characters to search.</CommandEmpty>
                  ) : isFetching ? (
                    <div className="flex items-center justify-center gap-2 py-6 text-sm text-muted">
                      <Loader2 className="size-4 animate-spin" />
                      Searching...
                    </div>
                  ) : results.length === 0 ? (
                    <CommandEmpty>No objects found.</CommandEmpty>
                  ) : (
                    <CommandGroup>
                      {results.map((obj, idx) => {
                        const name = typeof obj.name === "string" ? obj.name : "";
                        if (!name) return null;
                        const selected = value.includes(name);
                        const ip = firstText(obj, [
                          "ipv4-address",
                          "ipv6-address",
                          "subnet4",
                          "ipv4-address-first",
                        ]);
                        const comments = firstText(obj, ["comments"]);
                        return (
                          <CommandItem key={`${name}-${idx}`} value={name} onSelect={() => toggle(name)}>
                            <Check className={selected ? "size-4 shrink-0 text-accent" : "size-4 shrink-0 opacity-0"} />
                            <div className="flex min-w-0 flex-1 flex-col">
                              <span className="truncate">{name}</span>
                              {(ip || comments) && (
                                <span className="truncate text-xs text-muted">
                                  {[ip, comments].filter(Boolean).join(" · ")}
                                </span>
                              )}
                            </div>
                            {typeof obj.type === "string" && (
                              <span className="shrink-0 text-xs text-muted">{obj.type}</span>
                            )}
                          </CommandItem>
                        );
                      })}
                    </CommandGroup>
                  )}
                </CommandList>
              </Command>

              {creatable.length > 0 && (
                <div className="flex items-center justify-end border-t border-border p-2">
                  <Button type="button" variant="ghost" size="sm" className="gap-1.5 text-accent" onClick={openCreate}>
                    <Plus className="size-3.5" />
                    {creatable.length === 1
                      ? `Create new ${objectKinds[creatable[0]]?.title ?? creatable[0]}`
                      : "Create new object"}
                  </Button>
                </div>
              )}
            </>
          )}
        </PopoverContent>
      </Popover>

      {value.length > 0 && (
        <div className="flex flex-wrap gap-1.5">
          {value.map((name) => (
            <Badge key={name} variant="outline" className="gap-1 pr-1">
              {name}
              <button
                type="button"
                onClick={() => remove(name)}
                className="rounded-full p-0.5 hover:bg-surface-2"
                aria-label={`Remove ${name}`}
              >
                <X className="size-3" />
              </button>
            </Badge>
          ))}
        </div>
      )}
    </div>
  );
}
