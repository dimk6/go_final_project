package api

import (
"errors"
"fmt"
"strconv"
"strings"
"time"
)

const DateFormat = "20060102"

func parseList(s string) ([]int, error) {
parts := strings.Split(s, ",")
res := make([]int, 0, len(parts))
for _, p := range parts {
p = strings.TrimSpace(p)
if p == "" {
continue
}
val, err := strconv.Atoi(p)
if err != nil {
return nil, err
}
res = append(res, val)
}
return res, nil
}

func NextDate(now time.Time, date string, repeat string) (string, error) {
if repeat == "" {
return "", nil
}

startDate, err := time.Parse(DateFormat, date)
if err != nil {
return "", fmt.Errorf("invalid date: %w", err)
}

parts := strings.Fields(repeat)
if len(parts) == 0 {
return "", errors.New("empty repeat rule")
}

rule := parts[0]
switch rule {
case "y":
res := startDate
for {
res = res.AddDate(1, 0, 0)
if res.After(now) {
return res.Format(DateFormat), nil
}
}

case "d":
if len(parts) != 2 {
return "", errors.New("invalid format for rule d")
}
days, err := strconv.Atoi(parts[1])
if err != nil || days <= 0 || days > 400 {
return "", errors.New("invalid days count for rule d (1..400)")
}

res := startDate
for {
res = res.AddDate(0, 0, days)
if res.After(now) {
return res.Format(DateFormat), nil
}
}

case "w":
if len(parts) != 2 {
return "", errors.New("invalid format for rule w")
}
daysOfWeek, err := parseList(parts[1])
if err != nil || len(daysOfWeek) == 0 {
return "", errors.New("invalid days for rule w")
}
validDays := make(map[int]bool)
for _, d := range daysOfWeek {
if d < 1 || d > 7 {
return "", errors.New("invalid day of week: must be 1..7")
}
validDays[d] = true
}

cur := startDate
for {
cur = cur.AddDate(0, 0, 1)
weekday := int(cur.Weekday())
if weekday == 0 {
weekday = 7
}
if validDays[weekday] && cur.After(now) {
return cur.Format(DateFormat), nil
}
}

case "m":
if len(parts) < 2 || len(parts) > 3 {
return "", errors.New("invalid format for rule m")
}
days, err := parseList(parts[1])
if err != nil || len(days) == 0 {
return "", errors.New("invalid days for rule m")
}

var months map[int]bool
if len(parts) == 3 {
monthsList, err := parseList(parts[2])
if err != nil || len(monthsList) == 0 {
return "", errors.New("invalid months for rule m")
}
months = make(map[int]bool)
for _, m := range monthsList {
if m < 1 || m > 12 {
return "", errors.New("invalid month: must be 1..12")
}
months[m] = true
}
}

cur := startDate
for i := 0; i < 4000; i++ {
cur = cur.AddDate(0, 0, 1)
if months != nil && !months[int(cur.Month())] {
continue
}

day := cur.Day()
lastDayOfMonth := time.Date(cur.Year(), cur.Month()+1, 0, 0, 0, 0, 0, cur.Location()).Day()

match := false
for _, d := range days {
target := d
if d < 0 {
target = lastDayOfMonth + 1 + d
}
if target == day {
match = true
break
}
}

if match && cur.After(now) {
return cur.Format(DateFormat), nil
}
}
return "", errors.New("next date not found")

default:
return "", fmt.Errorf("unknown rule: %s", rule)
}
}
