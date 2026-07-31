import { useState, type ReactNode } from "react";
import { ChevronDown, ChevronRight } from "lucide-react";

import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { EmptyState } from "@/components/shared/EmptyState";
import type { JsonRecord } from "@/lib/wailsService";

export interface RulebaseColumn {
  header: string;
  cell: (row: JsonRecord) => ReactNode;
}

interface RulebaseTableProps {
  /** Flat rows from a show-*-rulebase call — a mix of rule rows and section
   * rows (rows whose `type` ends in "-section", e.g. "nat-section"). */
  rows: JsonRecord[];
  columns: RulebaseColumn[];
  /** Per-rule actions cell (Edit/Delete) — sections never get one. */
  renderActions?: (row: JsonRecord) => ReactNode;
  emptyMessage?: string;
}

function isSectionRow(row: JsonRecord): boolean {
  return typeof row.type === "string" && row.type.endsWith("-section");
}

/** Renders a rulebase (access/nat/threat/https) as a table, expanding
 * section rows in place to show their nested rules.
 *
 * The Management API nests a section's actual rules inside that section
 * row's own "rulebase" array rather than flattening them into the top-level
 * list — confirmed live against a lab where an "Automatic Generated Rules"
 * NAT section had rules nobody could see because nothing rendered that
 * nested array. Sections start collapsed and toggle open on click; rule
 * count is shown so an empty section (nothing configured yet, common for a
 * fresh policy) doesn't look broken. */
export function RulebaseTable({ rows, columns, renderActions, emptyMessage = "No rules." }: RulebaseTableProps) {
  const [expanded, setExpanded] = useState<Set<string>>(new Set());

  function toggle(uid: string) {
    setExpanded((prev) => {
      const next = new Set(prev);
      if (next.has(uid)) next.delete(uid);
      else next.add(uid);
      return next;
    });
  }

  if (rows.length === 0) {
    return <EmptyState title={emptyMessage} />;
  }

  const colSpan = columns.length + (renderActions ? 1 : 0);

  return (
    <Table>
      <TableHeader>
        <TableRow>
          {columns.map((col) => (
            <TableHead key={col.header}>{col.header}</TableHead>
          ))}
          {renderActions && <TableHead className="text-right">Actions</TableHead>}
        </TableRow>
      </TableHeader>
      <TableBody>
        {rows.map((row, idx) => {
          if (isSectionRow(row)) {
            const uid = typeof row.uid === "string" ? row.uid : `section-${idx}`;
            const isOpen = expanded.has(uid);
            const nested = Array.isArray(row.rulebase) ? (row.rulebase as JsonRecord[]) : [];
            return (
              <>
                <TableRow
                  key={uid}
                  className="cursor-pointer bg-surface-2/70 hover:bg-surface-2"
                  onClick={() => toggle(uid)}
                >
                  <TableCell colSpan={colSpan} className="font-medium text-muted">
                    <span className="flex items-center gap-1.5">
                      {isOpen ? <ChevronDown className="size-4" /> : <ChevronRight className="size-4" />}
                      {typeof row.name === "string" && row.name ? row.name : "Section"}
                      <span className="font-normal text-muted/70">
                        ({nested.length} {nested.length === 1 ? "rule" : "rules"})
                      </span>
                    </span>
                  </TableCell>
                </TableRow>
                {isOpen &&
                  (nested.length === 0 ? (
                    <TableRow key={`${uid}-empty`}>
                      <TableCell colSpan={colSpan} className="text-sm text-muted">
                        No rules in this section.
                      </TableCell>
                    </TableRow>
                  ) : (
                    nested.map((nestedRow, nestedIdx) => {
                      const nestedUid =
                        typeof nestedRow.uid === "string" ? nestedRow.uid : `${uid}-${nestedIdx}`;
                      return (
                        <TableRow key={nestedUid} className="bg-surface/60">
                          {columns.map((col) => (
                            <TableCell key={col.header}>{col.cell(nestedRow)}</TableCell>
                          ))}
                          {renderActions && (
                            <TableCell className="text-right">{renderActions(nestedRow)}</TableCell>
                          )}
                        </TableRow>
                      );
                    })
                  ))}
              </>
            );
          }

          const uid = typeof row.uid === "string" ? row.uid : String(idx);
          return (
            <TableRow key={uid}>
              {columns.map((col) => (
                <TableCell key={col.header}>{col.cell(row)}</TableCell>
              ))}
              {renderActions && <TableCell className="text-right">{renderActions(row)}</TableCell>}
            </TableRow>
          );
        })}
      </TableBody>
    </Table>
  );
}
