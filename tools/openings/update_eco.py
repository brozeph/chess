#!/usr/bin/env python3
import csv
import html
import re
import socket
import sys
import time
from pathlib import Path
from typing import List, Tuple
from urllib.error import HTTPError, URLError
from urllib.parse import urljoin
from urllib.request import Request, urlopen

BASE_URL = "https://www.365chess.com"
ECO_URL_TEMPLATE = BASE_URL + "/eco/{}"
PROJECT_ROOT = Path(__file__).resolve().parents[2]
OPENINGS_CSV = PROJECT_ROOT / "data" / "openings.csv"
USER_AGENT = "Mozilla/5.0 (compatible; CodexBot/1.0; +https://openai.com)"
REQUEST_TIMEOUT = 30
MAX_RETRIES = 3

MOVE_PAREN_RE = re.compile(r"\([^)]*\)")
NONMOVE_CHARS = str.maketrans(
    {
        ".": " ",
        "!": " ",
        "?": " ",
        "+": " ",
        "#": " ",
        ",": " ",
        ";": " ",
        ":": " ",
    }
)


class ECOPageParser:
    """Parses ECO pages that contain <div id="rel_ops2"> lists."""

    def __init__(self) -> None:
        from html.parser import HTMLParser

        class _Parser(HTMLParser):
            def __init__(self):
                super().__init__()
                self.entries: List[Tuple[str, str]] = []
                self._in_section = False
                self._section_depth = 0
                self._current_name = None
                self._current_moves = None
                self._reading_name = False
                self._reading_moves = False

            def handle_starttag(self, tag, attrs):
                attrs_dict = dict(attrs)
                if tag == "div" and attrs_dict.get("id") == "rel_ops2":
                    self._in_section = True
                    self._section_depth = 1
                    return

                if self._in_section and tag == "div":
                    self._section_depth += 1

                if not self._in_section:
                    return

                if tag == "li":
                    self._current_name = ""
                    self._current_moves = ""
                    self._reading_name = False
                    self._reading_moves = False
                elif tag == "a" and self._current_name is not None:
                    self._reading_name = True
                elif tag == "br" and self._current_name is not None:
                    self._reading_name = False
                    self._reading_moves = True

            def handle_endtag(self, tag):
                if self._in_section and tag == "li" and self._current_name is not None:
                    name = self._current_name.strip()
                    moves = (self._current_moves or "").strip()
                    if name and moves:
                        self.entries.append((name, moves))
                    self._current_name = None
                    self._current_moves = None
                    self._reading_name = False
                    self._reading_moves = False

                if tag == "a" and self._reading_name:
                    self._reading_name = False

                if self._in_section and tag == "div":
                    self._section_depth -= 1
                    if self._section_depth <= 0:
                        self._in_section = False

            def handle_data(self, data):
                if not self._in_section or self._current_name is None:
                    return
                if self._reading_name:
                    self._current_name += data
                elif self._reading_moves:
                    self._current_moves += data

        self._parser = _Parser()

    def parse(self, html_text: str) -> List[Tuple[str, str]]:
        self._parser.entries.clear()
        self._parser.feed(html_text)
        return [(n, m) for n, m in self._parser.entries]


def normalize_moves(moves: str) -> List[str]:
    """Normalizes a moves string into a list of tokens."""
    text = html.unescape(moves or "").replace("\xa0", " ")
    text = text.replace("O-O-O", "0-0-0").replace("O-O", "0-0")
    text = text.replace("–", "-").replace("—", "-")
    text = MOVE_PAREN_RE.sub("", text)
    text = text.translate(NONMOVE_CHARS)
    text = re.sub(r"\s+", " ", text).strip()
    return text.split() if text else []


def fetch_html(url: str) -> str:
    """Fetches HTML with retries."""
    last_err = None
    for attempt in range(1, MAX_RETRIES + 1):
        try:
            req = Request(url, headers={"User-Agent": USER_AGENT})
            with urlopen(req, timeout=REQUEST_TIMEOUT) as resp:
                return resp.read().decode("iso-8859-1", errors="ignore")
        except (HTTPError, URLError, TimeoutError, socket.timeout) as exc:
            last_err = exc
            time.sleep(1)
    raise RuntimeError(f"failed to fetch {url}: {last_err}")


def build_mapping() -> List[Tuple[str, List[str]]]:
    """Builds a list of (eco_code, token_list) tuples from the site."""
    parser = ECOPageParser()
    mapping: List[Tuple[str, List[str]]] = []
    codes = [f"{letter}{i:02d}" for letter in "ABCDE" for i in range(100)]

    for code in codes:
        url = ECO_URL_TEMPLATE.format(code)
        try:
            html_text = fetch_html(url)
        except RuntimeError as exc:
            print(exc, file=sys.stderr)
            continue
        entries = parser.parse(html_text)
        for name, moves in entries:
            tokens = normalize_moves(moves)
            if tokens:
                mapping.append((code, tokens))

    mapping.sort(key=lambda item: len(item[1]), reverse=True)
    return mapping


def find_code(tokens: List[str], mapping: List[Tuple[str, List[str]]]) -> str:
    """Finds the best matching ECO code for the given move tokens."""
    for code, seq in mapping:
        seq_len = len(seq)
        if seq_len <= len(tokens) and tokens[:seq_len] == seq:
            return code
    raise KeyError(f"No ECO match for moves: {' '.join(tokens)}")


def update_openings_csv(mapping: List[Tuple[str, List[str]]]) -> None:
    """Updates the CSV file with the ECO column."""
    with OPENINGS_CSV.open(newline="") as infile:
        reader = csv.DictReader(infile)
        rows = list(reader)

    for row in rows:
        tokens = normalize_moves(row["Moves"])
        try:
            row["ECO"] = find_code(tokens, mapping)
        except KeyError:
            row["ECO"] = ""
            print(f"warning: could not match moves '{row['Moves']}'", file=sys.stderr)

    fieldnames = ["Moves", "ECO", "Name", "ResultFEN", "SequenceFENs"]
    with OPENINGS_CSV.open("w", newline="") as outfile:
        writer = csv.DictWriter(outfile, fieldnames=fieldnames, lineterminator="\n")
        writer.writeheader()
        for row in rows:
            writer.writerow(
                {
                    "Moves": row["Moves"],
                    "ECO": row.get("ECO", ""),
                    "Name": row["Name"],
                    "ResultFEN": row["ResultFEN"],
                    "SequenceFENs": row["SequenceFENs"],
                }
            )


def main() -> None:
    mapping = build_mapping()
    if not mapping:
        print("No ECO entries scraped; aborting.", file=sys.stderr)
        sys.exit(1)
    update_openings_csv(mapping)
    print(f"Updated {OPENINGS_CSV} with ECO codes for {len(mapping)} entries.")


if __name__ == "__main__":
    main()
