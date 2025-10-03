'use client';

import { format } from 'date-fns';
import { CalendarIcon } from 'lucide-react';
import { useState } from 'react';
import type { DateRange } from 'react-day-picker';

import { cn } from '../../../lib/utils';
import { Button } from '../../ui/button';
import { Calendar } from '../../ui/calendar';
import { Popover, PopoverContent, PopoverTrigger } from '../../ui/popover';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../../ui/select';

export interface TimeRangeSelectorProps {
  onRangeChange?: (range: DateRange | undefined) => void;
  onPresetChange?: (preset: string) => void;
  showPresets?: boolean;
  className?: string;
}

const presets = [
  { value: '1h', label: 'Last hour' },
  { value: '24h', label: 'Last 24 hours' },
  { value: '7d', label: 'Last 7 days' },
  { value: '30d', label: 'Last 30 days' },
  { value: '90d', label: 'Last 90 days' },
  { value: 'custom', label: 'Custom range' },
];

export function TimeRangeSelector({
  onRangeChange,
  onPresetChange,
  showPresets = true,
  className,
}: TimeRangeSelectorProps) {
  const [date, setDate] = useState<DateRange | undefined>();
  const [selectedPreset, setSelectedPreset] = useState<string>('7d');

  const handlePresetChange = (value: string) => {
    setSelectedPreset(value);
    if (value === 'custom') {
      return;
    }

    const now = new Date();
    const from = new Date();

    switch (value) {
      case '1h':
        from.setHours(from.getHours() - 1);
        break;
      case '24h':
        from.setDate(from.getDate() - 1);
        break;
      case '7d':
        from.setDate(from.getDate() - 7);
        break;
      case '30d':
        from.setDate(from.getDate() - 30);
        break;
      case '90d':
        from.setDate(from.getDate() - 90);
        break;
    }

    const range = { from, to: now };
    setDate(range);
    onRangeChange?.(range);
    onPresetChange?.(value);
  };

  const handleDateChange = (newDate: DateRange | undefined) => {
    setDate(newDate);
    onRangeChange?.(newDate);
    if (newDate) {
      setSelectedPreset('custom');
      onPresetChange?.('custom');
    }
  };

  return (
    <div className={cn('flex items-center gap-2', className)}>
      {showPresets && (
        <Select value={selectedPreset} onValueChange={handlePresetChange}>
          <SelectTrigger className="w-40">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {presets.map((preset) => (
              <SelectItem key={preset.value} value={preset.value}>
                {preset.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      )}

      {selectedPreset === 'custom' && (
        <Popover>
          <PopoverTrigger asChild>
            <Button
              variant="outline"
              className={cn(
                'justify-start text-left font-normal',
                !date && 'text-muted-foreground'
              )}
            >
              <CalendarIcon className="mr-2 h-4 w-4" />
              {date?.from ? (
                date.to ? (
                  <>
                    {format(date.from, 'LLL dd, y')} - {format(date.to, 'LLL dd, y')}
                  </>
                ) : (
                  format(date.from, 'LLL dd, y')
                )
              ) : (
                <span>Pick a date range</span>
              )}
            </Button>
          </PopoverTrigger>
          <PopoverContent className="w-auto p-0" align="start">
            <Calendar
              initialFocus
              mode="range"
              defaultMonth={date?.from}
              selected={date}
              onSelect={handleDateChange}
              numberOfMonths={2}
            />
          </PopoverContent>
        </Popover>
      )}
    </div>
  );
}
