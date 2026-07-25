import { useEffect, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Check, ChevronDown, Loader2, X } from "lucide-react";

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
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { getService } from "@/lib/wailsService";

interface ObjectPickerProps {
  /** Selected object names. */
  value: string[];
  onChange: (names: string[]) => void;
  placeholder?: string;
  /** Narrows the search to one Check Point object type (e.g. "host") when
   * set. Left blank searches across every type — the API only accepts one
   * type at a time, and fields like source/destination legitimately accept
   * several (host, network, group, address-range, ...), so most callers
   * leave this unset. */
  objType?: string;
}

function useDebouncedValue(value: string, delayMs: number): string {
  const [debounced, setDebounced] = useState(value);
  useEffect(() => {
    const id = setTimeout(() => setDebounced(value), delayMs);
    return () => clearTimeout(id);
  }, [value, delayMs]);
  return debounced;
}

/** Search-and-select combobox for Check Point objects (hosts, networks,
 * groups, services, ...) — replaces free-text "type the exact name"
 * inputs on Origem/Destino/Serviço-style fields with the same live-search
 * flow SmartConsole's own object picker uses (via the Management API's
 * `show-objects` with a text filter, `service.Service.SearchObjects`).
 *
 * v1 scope: plain text search across all types, no category sidebar/counts
 * (SmartConsole shows "Network Objects 17", "Services 521", etc.) and no
 * per-field type narrowing beyond the optional `objType` prop — the API
 * only takes one type per search and fields like source/destination accept
 * several valid types, so full category filtering is a follow-up, not done
 * here. */
export function ObjectPicker({ value, onChange, placeholder = "Buscar objetos...", objType }: ObjectPickerProps) {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const debouncedQuery = useDebouncedValue(query, 250);

  const { data = [], isFetching } = useQuery({
    queryKey: ["object-search", debouncedQuery, objType ?? ""],
    queryFn: () => getService().searchObjects(debouncedQuery, objType ?? ""),
    enabled: open && debouncedQuery.trim().length >= 2,
  });

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

  return (
    <div className="flex flex-col gap-1.5">
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            type="button"
            variant="outline"
            role="combobox"
            aria-expanded={open}
            className="h-9 w-full justify-between font-normal"
          >
            <span className="truncate text-muted">
              {value.length > 0 ? `${value.length} selecionado${value.length === 1 ? "" : "s"}` : placeholder}
            </span>
            <ChevronDown className="size-4 shrink-0 opacity-60" />
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-[320px]">
          <Command shouldFilter={false}>
            <CommandInput placeholder="Digite ao menos 2 letras..." value={query} onValueChange={setQuery} />
            <CommandList>
              {query.trim().length < 2 ? (
                <CommandEmpty>Digite ao menos 2 letras para buscar.</CommandEmpty>
              ) : isFetching ? (
                <div className="flex items-center justify-center gap-2 py-6 text-sm text-muted">
                  <Loader2 className="size-4 animate-spin" />
                  Buscando...
                </div>
              ) : data.length === 0 ? (
                <CommandEmpty>Nenhum objeto encontrado.</CommandEmpty>
              ) : (
                <CommandGroup>
                  {data.map((obj, idx) => {
                    const name = typeof obj.name === "string" ? obj.name : "";
                    if (!name) return null;
                    const selected = value.includes(name);
                    return (
                      <CommandItem key={`${name}-${idx}`} value={name} onSelect={() => toggle(name)}>
                        <Check className={selected ? "size-4 text-accent" : "size-4 opacity-0"} />
                        <span className="flex-1 truncate">{name}</span>
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
                aria-label={`Remover ${name}`}
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
