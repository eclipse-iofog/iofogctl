/*
 *  *******************************************************************************
 *  * Copyright (c) 2019 Edgeworx, Inc.
 *  *
 *  * This program and the accompanying materials are made available under the
 *  * terms of the Eclipse Public License v. 2.0 which is available at
 *  * http://www.eclipse.org/legal/epl-2.0
 *  *
 *  * SPDX-License-Identifier: EPL-2.0
 *  *******************************************************************************
 *
 */

package util

import (
	"fmt"
	"time"
)

var timeFormat = time.UnixDate

func Now() string {
	return time.Now().Format(timeFormat)
}

func FromInt(sec int64) string {
	return time.Unix(sec, 0).Format(timeFormat)
}

func Elapsed(from, to string) (diff string, err error) {
	fromTime, err := time.Parse(timeFormat, from)
	if err != nil {
		return
	}
	diffTime := time.Now().Sub(fromTime)
	diff = FormatDuration(diffTime)
	return
}

func FormatDuration(duration time.Duration) string {
	duration = duration.Round(time.Second)
	// Get days
	days := duration / (time.Hour * 24)
	duration -= days * (time.Hour * 24)

	// Get hours
	hours := duration / time.Hour
	duration -= hours * time.Hour

	// Get Minutes
	mins := duration / time.Minute
	duration -= mins * time.Minute

	// Get Seconds
	secs := duration / time.Second

	if days > 0 {
		return fmt.Sprintf("%dd%dh", days, hours)
	}

	if hours > 0 {
		return fmt.Sprintf("%dh%dm", hours, mins)
	}

	if mins > 0 {
		return fmt.Sprintf("%dm%ds", mins, secs)
	}

	return fmt.Sprintf("%ds", secs)
}
