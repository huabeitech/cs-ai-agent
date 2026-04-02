"use client";

import { DownloadIcon, FileUpIcon, InfoIcon } from "lucide-react";
import { useMemo, useRef, useState } from "react";
import { toast } from "sonner";

import { ProjectDialog } from "@/components/project-dialog";
import { Button } from "@/components/ui/button";
import {
  Field,
  FieldContent,
  FieldDescription,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { createKnowledgeFAQ, type CreateKnowledgeFAQPayload } from "@/lib/api/admin";

type FAQImportDialogProps = {
  open: boolean;
  knowledgeBaseId: number | null;
  importing: boolean;
  onOpenChange: (open: boolean) => void;
  onImportingChange: (importing: boolean) => void;
  onImported: () => Promise<void>;
};

type ParsedFAQRow = {
  rowNo: number;
  question: string;
  answer: string;
  similarQuestions: string[];
  remark: string;
};

type ParseResult = {
  rows: ParsedFAQRow[];
  warnings: string[];
};

const templateContent = [
  "question,answer,similarQuestions,remark",
  '"如何重置密码?","进入个人设置后点击重置密码。","忘记密码|密码重置在哪里","账号类FAQ"',
  '"支持哪些接入渠道?","目前支持网站组件、企业微信等渠道接入。","有哪些渠道|支持什么渠道","渠道说明"',
].join("\n");

const acceptedHeaderMap: Record<string, keyof Omit<ParsedFAQRow, "rowNo">> = {
  question: "question",
  "标准问题": "question",
  "问题": "question",
  answer: "answer",
  "答案": "answer",
  similarquestions: "similarQuestions",
  "similarQuestions": "similarQuestions",
  "相似问": "similarQuestions",
  "相似问题": "similarQuestions",
  remark: "remark",
  "备注": "remark",
};

function normalizeHeader(value: string) {
  return value.trim().replace(/^\uFEFF/, "").toLowerCase();
}

function parseDelimitedText(input: string): string[][] {
  const text = input.replace(/\r\n/g, "\n").replace(/\r/g, "\n");
  const rows: string[][] = [];
  let row: string[] = [];
  let cell = "";
  let inQuotes = false;
  let delimiter = ",";

  function pushCell() {
    row.push(cell.trim());
    cell = "";
  }

  function pushRow() {
    if (row.length === 1 && row[0] === "" && rows.length === 0) {
      row = [];
      return;
    }
    if (row.some((item) => item !== "")) {
      rows.push(row);
    }
    row = [];
  }

  for (let i = 0; i < text.length; i += 1) {
    const char = text[i];
    const next = text[i + 1];

    if (!inQuotes && rows.length === 0 && row.length === 0 && cell.length > 0 && char === "\t") {
      delimiter = "\t";
    }

    if (char === '"') {
      if (inQuotes && next === '"') {
        cell += '"';
        i += 1;
        continue;
      }
      inQuotes = !inQuotes;
      continue;
    }

    if (!inQuotes && char === delimiter) {
      pushCell();
      continue;
    }

    if (!inQuotes && char === "\n") {
      pushCell();
      pushRow();
      continue;
    }

    cell += char;
  }

  if (cell.length > 0 || row.length > 0) {
    pushCell();
    pushRow();
  }

  return rows;
}

function parseSimilarQuestions(value: string) {
  return value
    .split(/\r?\n|\|/g)
    .map((item) => item.trim())
    .filter(Boolean);
}

function parseFAQFileContent(input: string): ParseResult {
  const table = parseDelimitedText(input);
  if (table.length === 0) {
    throw new Error("文件内容为空");
  }

  const headerRow = table[0];
  const headerMap = new Map<keyof Omit<ParsedFAQRow, "rowNo">, number>();
  for (let index = 0; index < headerRow.length; index += 1) {
    const header = acceptedHeaderMap[normalizeHeader(headerRow[index])];
    if (header && !headerMap.has(header)) {
      headerMap.set(header, index);
    }
  }

  if (!headerMap.has("question") || !headerMap.has("answer")) {
    throw new Error("导入模板缺少 question/answer 列");
  }

  const rows: ParsedFAQRow[] = [];
  const warnings: string[] = [];

  for (let index = 1; index < table.length; index += 1) {
    const current = table[index];
    const rowNo = index + 1;
    const question = current[headerMap.get("question") ?? -1]?.trim() ?? "";
    const answer = current[headerMap.get("answer") ?? -1]?.trim() ?? "";
    const similarQuestionsRaw = current[headerMap.get("similarQuestions") ?? -1]?.trim() ?? "";
    const remark = current[headerMap.get("remark") ?? -1]?.trim() ?? "";

    if (!question && !answer && !similarQuestionsRaw && !remark) {
      continue;
    }
    if (!question || !answer) {
      warnings.push(`第 ${rowNo} 行缺少问题或答案，已跳过`);
      continue;
    }

    rows.push({
      rowNo,
      question,
      answer,
      similarQuestions: parseSimilarQuestions(similarQuestionsRaw),
      remark,
    });
  }

  return { rows, warnings };
}

function downloadTemplate() {
  const blob = new Blob([templateContent], { type: "text/csv;charset=utf-8;" });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = "knowledge-faq-import-template.csv";
  link.click();
  URL.revokeObjectURL(url);
}

function buildPayload(row: ParsedFAQRow, knowledgeBaseId: number): CreateKnowledgeFAQPayload {
  return {
    knowledgeBaseId,
    question: row.question,
    answer: row.answer,
    similarQuestions: row.similarQuestions,
    remark: row.remark,
  };
}

export function FAQImportDialog({
  open,
  knowledgeBaseId,
  importing,
  onOpenChange,
  onImportingChange,
  onImported,
}: FAQImportDialogProps) {
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const [fileName, setFileName] = useState("");
  const [rows, setRows] = useState<ParsedFAQRow[]>([]);
  const [warnings, setWarnings] = useState<string[]>([]);

  const previewRows = useMemo(() => rows.slice(0, 5), [rows]);

  function resetState() {
    setFileName("");
    setRows([]);
    setWarnings([]);
    if (fileInputRef.current) {
      fileInputRef.current.value = "";
    }
  }

  async function handleFileChange(event: React.ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0];
    if (!file) {
      return;
    }

    try {
      const content = await file.text();
      const parsed = parseFAQFileContent(content);
      setFileName(file.name);
      setRows(parsed.rows);
      setWarnings(parsed.warnings);
      if (parsed.rows.length === 0) {
        toast.error("没有可导入的FAQ记录");
      } else {
        toast.success(`已解析 ${parsed.rows.length} 条FAQ`);
      }
    } catch (error) {
      resetState();
      toast.error(error instanceof Error ? error.message : "解析导入文件失败");
    }
  }

  async function handleImport() {
    if (!knowledgeBaseId || rows.length === 0 || importing) {
      return;
    }

    onImportingChange(true);
    let successCount = 0;
    const failedRows: string[] = [];

    try {
      for (const row of rows) {
        try {
          await createKnowledgeFAQ(buildPayload(row, knowledgeBaseId));
          successCount += 1;
        } catch (error) {
          failedRows.push(
            `第 ${row.rowNo} 行：${error instanceof Error ? error.message : "导入失败"}`
          );
        }
      }

      await onImported();

      if (successCount > 0) {
        toast.success(`成功导入 ${successCount} 条FAQ`);
      }
      if (failedRows.length > 0) {
        toast.error(`有 ${failedRows.length} 条FAQ导入失败`);
        setWarnings((current) => [...current, ...failedRows]);
        return;
      }

      resetState();
      onOpenChange(false);
    } finally {
      onImportingChange(false);
    }
  }

  return (
    <ProjectDialog
      open={open}
      onOpenChange={(nextOpen) => {
        if (!nextOpen && !importing) {
          resetState();
        }
        onOpenChange(nextOpen);
      }}
      title="导入FAQ"
      description="上传 CSV 文件批量导入 FAQ。必填列为 question、answer；similarQuestions 使用 | 或换行分隔。"
      size="lg"
      footer={
        <>
          <Button type="button" variant="outline" onClick={() => downloadTemplate()}>
            <DownloadIcon className="size-4" />
            下载模板
          </Button>
          <Button type="button" variant="outline" onClick={() => onOpenChange(false)} disabled={importing}>
            取消
          </Button>
          <Button type="button" onClick={() => void handleImport()} disabled={importing || rows.length === 0}>
            {importing ? "导入中..." : `开始导入${rows.length > 0 ? ` (${rows.length})` : ""}`}
          </Button>
        </>
      }
    >
      <FieldGroup>
        <Field>
          <FieldLabel htmlFor="faq-import-file">导入文件</FieldLabel>
          <FieldContent>
            <div className="flex items-center gap-2">
              <Input
                id="faq-import-file"
                ref={fileInputRef}
                type="file"
                accept=".csv,text/csv,.txt"
                onChange={(event) => void handleFileChange(event)}
              />
              <Button
                type="button"
                variant="outline"
                onClick={() => fileInputRef.current?.click()}
              >
                <FileUpIcon className="size-4" />
                选择文件
              </Button>
            </div>
            <FieldDescription>
              支持 UTF-8 编码的 CSV 或制表符文本文件。
            </FieldDescription>
          </FieldContent>
        </Field>

        {fileName ? (
          <div className="rounded-md border bg-muted/20 px-3 py-2 text-sm">
            当前文件：{fileName}
          </div>
        ) : null}

        {warnings.length > 0 ? (
          <div className="rounded-md border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-900">
            <div className="mb-2 flex items-center gap-2 font-medium">
              <InfoIcon className="size-4" />
              导入提示
            </div>
            <ul className="space-y-1">
              {warnings.map((item, index) => (
                <li key={`${item}-${index}`}>{item}</li>
              ))}
            </ul>
          </div>
        ) : null}

        <div className="rounded-md border">
          <div className="border-b px-4 py-3 text-sm font-medium">
            导入预览
          </div>
          {previewRows.length > 0 ? (
            <ScrollArea className="max-h-80">
              <div className="divide-y">
                {previewRows.map((row) => (
                  <div key={row.rowNo} className="space-y-2 px-4 py-3 text-sm">
                    <div>
                      <span className="text-muted-foreground">第 {row.rowNo} 行</span>
                    </div>
                    <div>
                      <div className="font-medium">{row.question}</div>
                      <div className="mt-1 whitespace-pre-wrap text-muted-foreground">
                        {row.answer}
                      </div>
                    </div>
                    <div className="text-muted-foreground">
                      相似问：{row.similarQuestions.length > 0 ? row.similarQuestions.join(" / ") : "无"}
                    </div>
                    <div className="text-muted-foreground">备注：{row.remark || "无"}</div>
                  </div>
                ))}
              </div>
            </ScrollArea>
          ) : (
            <div className="px-4 py-12 text-center text-sm text-muted-foreground">
              上传文件后可预览前 5 条FAQ
            </div>
          )}
        </div>
      </FieldGroup>
    </ProjectDialog>
  );
}
