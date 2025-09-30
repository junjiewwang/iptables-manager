/**
 * useTagTypes composable 单元测试
 */

import { describe, it, expect } from 'vitest'
import { useTagTypes } from '../useTagTypes'

describe('useTagTypes', () => {
  const { getTableTagType, getTargetTagType, getChainTagType, getTablesInGroup } = useTagTypes()

  describe('getTableTagType', () => {
    it('should return correct type for known table names', () => {
      expect(getTableTagType('raw')).toBe('info')
      expect(getTableTagType('mangle')).toBe('warning')
      expect(getTableTagType('nat')).toBe('success')
      expect(getTableTagType('filter')).toBe('danger')
    })

    it('should return default type for unknown table names', () => {
      expect(getTableTagType('unknown')).toBe('default')
      expect(getTableTagType('')).toBe('default')
    })
  })

  describe('getTargetTagType', () => {
    it('should return correct type for known targets', () => {
      expect(getTargetTagType('ACCEPT')).toBe('success')
      expect(getTargetTagType('DROP')).toBe('danger')
      expect(getTargetTagType('REJECT')).toBe('warning')
      expect(getTargetTagType('RETURN')).toBe('info')
      expect(getTargetTagType('MASQUERADE')).toBe('primary')
      expect(getTargetTagType('SNAT')).toBe('primary')
      expect(getTargetTagType('DNAT')).toBe('primary')
    })

    it('should return default type for unknown targets', () => {
      expect(getTargetTagType('UNKNOWN')).toBe('default')
      expect(getTargetTagType('')).toBe('default')
    })
  })

  describe('getChainTagType', () => {
    it('should return correct type for known chain names', () => {
      expect(getChainTagType('PREROUTING')).toBe('primary')
      expect(getChainTagType('INPUT')).toBe('success')
      expect(getChainTagType('FORWARD')).toBe('warning')
      expect(getChainTagType('OUTPUT')).toBe('info')
      expect(getChainTagType('POSTROUTING')).toBe('danger')
    })

    it('should return default type for unknown chain names', () => {
      expect(getChainTagType('UNKNOWN')).toBe('default')
      expect(getChainTagType('')).toBe('default')
    })
  })

  describe('getTablesInGroup', () => {
    it('should extract unique table names from rules', () => {
      const rules = [
        { table: 'filter', target: 'ACCEPT' },
        { table: 'nat', target: 'MASQUERADE' },
        { table: 'filter', target: 'DROP' },
        { table: 'mangle', target: 'ACCEPT' }
      ]

      const tables = getTablesInGroup(rules)
      expect(tables).toEqual(['filter', 'nat', 'mangle'])
    })

    it('should handle empty rules array', () => {
      const tables = getTablesInGroup([])
      expect(tables).toEqual([])
    })

    it('should filter out falsy table values', () => {
      const rules = [
        { table: 'filter', target: 'ACCEPT' },
        { table: null, target: 'DROP' },
        { table: '', target: 'REJECT' },
        { table: 'nat', target: 'MASQUERADE' }
      ]

      const tables = getTablesInGroup(rules)
      expect(tables).toEqual(['filter', 'nat'])
    })
  })
})