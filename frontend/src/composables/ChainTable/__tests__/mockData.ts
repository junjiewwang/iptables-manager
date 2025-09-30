import type { ChainTableData, NetworkInterface } from '@/types/ChainTable'

/**
 * 测试用模拟链表数据
 */
export const getMockChainTableData = (): ChainTableData => {
    return {
        chains: [
            {
                name: 'PREROUTING',
                tables: ['raw', 'mangle', 'nat'],
                rules: [
                    {
                        id: 1,
                        chain_name: 'PREROUTING',
                        table: 'mangle',
                        rule_text: 'MARK --set-mark 1',
                        line_number: '1',
                        target: 'MARK',
                        source: '0.0.0.0/0',
                        destination: '0.0.0.0/0',
                        protocol: 'all',
                        in_interface: 'any',
                        out_interface: 'any',
                        options: '--set-mark 1'
                    },
                    {
                        id: 2,
                        chain_name: 'PREROUTING',
                        table: 'nat',
                        rule_text: 'DNAT --to-destination 192.168.1.100',
                        line_number: '2',
                        target: 'DNAT',
                        source: '0.0.0.0/0',
                        destination: '0.0.0.0/0',
                        protocol: 'tcp',
                        in_interface: 'eth0',
                        out_interface: 'any',
                        options: '--to-destination 192.168.1.100'
                    }
                ]
            },
            {
                name: 'INPUT',
                tables: ['mangle', 'filter'],
                rules: [
                    {
                        id: 3,
                        chain_name: 'INPUT',
                        table: 'filter',
                        rule_text: 'ACCEPT -p tcp --dport 22',
                        line_number: '3',
                        target: 'ACCEPT',
                        source: '0.0.0.0/0',
                        destination: '0.0.0.0/0',
                        protocol: 'tcp',
                        in_interface: 'any',
                        out_interface: 'any',
                        options: '--dport 22'
                    },
                    {
                        id: 4,
                        chain_name: 'INPUT',
                        table: 'filter',
                        rule_text: 'ACCEPT -p tcp --dport 80',
                        line_number: '4',
                        target: 'ACCEPT',
                        source: '0.0.0.0/0',
                        destination: '0.0.0.0/0',
                        protocol: 'tcp',
                        in_interface: 'any',
                        out_interface: 'any',
                        options: '--dport 80'
                    },
                    {
                        id: 5,
                        chain_name: 'INPUT',
                        table: 'filter',
                        rule_text: 'DROP -p tcp --dport 23',
                        line_number: '5',
                        target: 'DROP',
                        source: '0.0.0.0/0',
                        destination: '0.0.0.0/0',
                        protocol: 'tcp',
                        in_interface: 'any',
                        out_interface: 'any',
                        options: '--dport 23'
                    }
                ]
            },
            {
                name: 'FORWARD',
                tables: ['mangle', 'filter'],
                rules: [
                    { 
                        id: 6, 
                        chain_name: 'FORWARD', 
                        table: 'filter', 
                        rule_text: 'ACCEPT -i eth0 -o eth1',
                        in_interface: 'eth0',
                        out_interface: 'eth1'
                    },
                    { 
                        id: 7, 
                        chain_name: 'FORWARD', 
                        table: 'filter', 
                        rule_text: 'DOCKER-ISOLATION-STAGE-1',
                        in_interface: 'any',
                        out_interface: 'any'
                    }
                ]
            },
            {
                name: 'OUTPUT',
                tables: ['raw', 'mangle', 'nat', 'filter'],
                rules: [
                    { 
                        id: 8, 
                        chain_name: 'OUTPUT', 
                        table: 'filter', 
                        rule_text: 'ACCEPT -p tcp --dport 443',
                        in_interface: 'any',
                        out_interface: 'any'
                    }
                ]
            },
            {
                name: 'POSTROUTING',
                tables: ['mangle', 'nat'],
                rules: [
                    { 
                        id: 9, 
                        chain_name: 'POSTROUTING', 
                        table: 'nat', 
                        rule_text: 'MASQUERADE -o eth0',
                        in_interface: 'any',
                        out_interface: 'eth0'
                    },
                    { 
                        id: 10, 
                        chain_name: 'POSTROUTING', 
                        table: 'nat', 
                        rule_text: 'SNAT --to-source 192.168.1.1',
                        in_interface: 'any',
                        out_interface: 'any'
                    }
                ]
            }
        ],
        tables: [
            {
                name: 'raw',
                total_rules: 0,
                chains: []
            },
            {
                name: 'mangle',
                total_rules: 1,
                chains: [
                    {
                        name: 'PREROUTING',
                        policy: 'ACCEPT',
                        rules: [{ id: 1, rule_text: 'MARK --set-mark 1' }]
                    }
                ]
            },
            {
                name: 'nat',
                total_rules: 3,
                chains: [
                    {
                        name: 'PREROUTING',
                        policy: 'ACCEPT',
                        rules: [{ id: 2, rule_text: 'DNAT --to-destination 192.168.1.100' }]
                    },
                    {
                        name: 'POSTROUTING',
                        policy: 'ACCEPT',
                        rules: [
                            { id: 9, rule_text: 'MASQUERADE -o eth0' },
                            { id: 10, rule_text: 'SNAT --to-source 192.168.1.1' }
                        ]
                    }
                ]
            },
            {
                name: 'filter',
                total_rules: 6,
                chains: [
                    {
                        name: 'INPUT',
                        policy: 'ACCEPT',
                        rules: [
                            { id: 3, rule_text: 'ACCEPT -p tcp --dport 22' },
                            { id: 4, rule_text: 'ACCEPT -p tcp --dport 80' },
                            { id: 5, rule_text: 'DROP -p tcp --dport 23' }
                        ]
                    },
                    {
                        name: 'FORWARD',
                        policy: 'ACCEPT',
                        rules: [
                            { id: 6, rule_text: 'ACCEPT -i eth0 -o eth1' },
                            { id: 7, rule_text: 'DOCKER-ISOLATION-STAGE-1' }
                        ]
                    },
                    {
                        name: 'OUTPUT',
                        policy: 'ACCEPT',
                        rules: [
                            { id: 8, rule_text: 'ACCEPT -p tcp --dport 443' }
                        ]
                    }
                ]
            }
        ],
        interfaceRules: {
            'eth0': [
                { id: 6, in_interface: 'eth0', rule_text: 'ACCEPT -i eth0 -o eth1' },
                { id: 9, out_interface: 'eth0', rule_text: 'MASQUERADE -o eth0' }
            ],
            'eth1': [
                { id: 6, out_interface: 'eth1', rule_text: 'ACCEPT -i eth0 -o eth1' }
            ]
        }
    }
}

/**
 * 测试用模拟网络接口数据
 */
export const getMockInterfaces = (): NetworkInterface[] => {
    return [
        {
            name: 'eth0',
            type: 'ethernet',
            state: 'UP',
            ip_addresses: ['192.168.1.100', '10.0.0.1'],
            mac_address: '00:1B:44:11:3A:B7',
            mtu: 1500,
            is_up: true,
            is_docker: false,
            statistics: {
                rx_bytes: 1024000,
                tx_bytes: 512000,
                rx_packets: 1000,
                tx_packets: 500
            }
        },
        {
            name: 'eth1',
            type: 'ethernet',
            state: 'UP',
            ip_addresses: ['172.16.0.1'],
            mac_address: '00:1B:44:11:3A:B8',
            mtu: 1500,
            is_up: true,
            is_docker: false,
            statistics: {
                rx_bytes: 2048000,
                tx_bytes: 1024000,
                rx_packets: 2000,
                tx_packets: 1000
            }
        },
        {
            name: 'docker0',
            type: 'bridge',
            state: 'UP',
            ip_addresses: ['172.17.0.1'],
            mac_address: '02:42:C0:A8:01:01',
            mtu: 1500,
            is_up: true,
            is_docker: true,
            statistics: {
                rx_bytes: 512000,
                tx_bytes: 256000,
                rx_packets: 500,
                tx_packets: 250
            }
        },
        {
            name: 'lo',
            type: 'loopback',
            state: 'UP',
            ip_addresses: ['127.0.0.1', '::1'],
            mac_address: '',
            mtu: 65536,
            is_up: true,
            is_docker: false,
            statistics: {
                rx_bytes: 1000000,
                tx_bytes: 1000000,
                rx_packets: 10000,
                tx_packets: 10000
            }
        }
    ]
}

/**
 * 测试用简化的链表数据
 */
export const getSimpleMockChainTableData = (): ChainTableData => {
    return {
        chains: [
            {
                name: 'INPUT',
                tables: ['filter'],
                rules: [
                    {
                        id: 1,
                        chain_name: 'INPUT',
                        table: 'filter',
                        rule_text: 'ACCEPT -p tcp --dport 22',
                        line_number: '1',
                        target: 'ACCEPT',
                        source: '0.0.0.0/0',
                        destination: '0.0.0.0/0',
                        protocol: 'tcp',
                        in_interface: 'any',
                        out_interface: 'any',
                        options: '--dport 22'
                    }
                ]
            }
        ],
        tables: [
            {
                name: 'filter',
                total_rules: 1,
                chains: [
                    {
                        name: 'INPUT',
                        policy: 'ACCEPT',
                        rules: [
                            { id: 1, rule_text: 'ACCEPT -p tcp --dport 22' }
                        ]
                    }
                ]
            }
        ]
    }
}