import React, { useState } from 'react'
import { ServiceContract } from '../types'
import { X, Layers, Plus, Save, Globe, Code2 } from 'lucide-react'

interface ContractEditorModalProps {
  contracts: ServiceContract[]
  isOpen: boolean
  onClose: () => void
  onSaveContract: (contract: ServiceContract) => Promise<void>
}

export const ContractEditorModal: React.FC<ContractEditorModalProps> = ({
  contracts,
  isOpen,
  onClose,
  onSaveContract,
}) => {
  const [selectedId, setSelectedId] = useState<string>(contracts[0]?.id || '')
  const [contractId, setContractId] = useState<string>('')
  const [contractName, setContractName] = useState<string>('')
  const [endpoint, setEndpoint] = useState<string>('')
  const [description, setDescription] = useState<string>('')
  const [fieldsJson, setFieldsJson] = useState<string>('{}')
  const [isSaving, setIsSaving] = useState<boolean>(false)
  const [error, setError] = useState<string | null>(null)

  React.useEffect(() => {
    if (selectedId) {
      const c = contracts.find((item) => item.id === selectedId)
      if (c) {
        setContractId(c.id)
        setContractName(c.name)
        setEndpoint(c.endpoint)
        setDescription(c.description)
        setFieldsJson(JSON.stringify(c.required_fields, null, 2))
      }
    } else if (contracts.length > 0) {
      setSelectedId(contracts[0].id)
    }
  }, [selectedId, contracts])

  if (!isOpen) return null

  const handleAddNew = () => {
    setSelectedId('')
    setContractId('new-service')
    setContractName('New Custom Microservice')
    setEndpoint('https://httpbin.org/post')
    setDescription('Custom microservice contract definition')
    setFieldsJson(JSON.stringify({ order_id: 'string', amount_cents: 'int' }, null, 2))
    setError(null)
  }

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      const parsedFields = JSON.parse(fieldsJson)
      setIsSaving(true)
      setError(null)
      await onSaveContract({
        id: contractId.trim(),
        name: contractName.trim(),
        endpoint: endpoint.trim(),
        description: description.trim(),
        required_fields: parsedFields,
      })
      setIsSaving(false)
      onClose()
    } catch (err: any) {
      setError(err.message)
      setIsSaving(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-fade-in">
      <div className="w-full max-w-2xl bg-canvas-cardDark border border-white/10 rounded-2xl p-6 shadow-2xl flex flex-col max-h-[90vh]">
        <div className="flex items-center justify-between pb-4 border-b border-white/10">
          <div className="flex items-center gap-2">
            <Layers className="w-5 h-5 text-blue-400" />
            <h3 className="text-base font-semibold text-white">
              Target Microservice Schema Contracts
            </h3>
          </div>
          <button
            onClick={onClose}
            className="p-1 rounded-lg text-slate-400 hover:text-white hover:bg-white/10 transition"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Existing Contracts Selector & New button */}
        <div className="flex items-center gap-2 my-4">
          <select
            value={selectedId}
            onChange={(e) => setSelectedId(e.target.value)}
            className="flex-1 bg-canvas-subtleDark border border-white/10 rounded-lg px-3 py-2 text-xs font-mono text-white focus:outline-none focus:border-brand-orange"
          >
            {contracts.map((c) => (
              <option key={c.id} value={c.id}>
                {c.name} ({c.id})
              </option>
            ))}
          </select>
          <button
            type="button"
            onClick={handleAddNew}
            className="flex items-center gap-1 px-3 py-2 rounded-lg text-xs font-semibold bg-white/10 text-white hover:bg-white/20 transition cursor-pointer"
          >
            <Plus className="w-3.5 h-3.5" />
            New Contract
          </button>
        </div>

        {/* Form */}
        <form onSubmit={handleSave} className="space-y-3.5 overflow-auto flex-1 pr-1">
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-xs font-mono text-slate-400 mb-1">
                Contract ID (URL Slug):
              </label>
              <input
                type="text"
                value={contractId}
                onChange={(e) => setContractId(e.target.value)}
                required
                className="w-full bg-canvas-subtleDark border border-white/10 rounded-lg px-3 py-1.5 text-xs text-white font-mono focus:border-brand-orange focus:outline-none"
              />
            </div>
            <div>
              <label className="block text-xs font-mono text-slate-400 mb-1">
                Service Display Name:
              </label>
              <input
                type="text"
                value={contractName}
                onChange={(e) => setContractName(e.target.value)}
                required
                className="w-full bg-canvas-subtleDark border border-white/10 rounded-lg px-3 py-1.5 text-xs text-white font-mono focus:border-brand-orange focus:outline-none"
              />
            </div>
          </div>

          <div>
            <label className="block text-xs font-mono text-slate-400 mb-1 flex items-center gap-1">
              <Globe className="w-3.5 h-3.5 text-cyan-400" />
              Target Microservice Upstream Endpoint:
            </label>
            <input
              type="text"
              value={endpoint}
              onChange={(e) => setEndpoint(e.target.value)}
              placeholder="e.g. internal://checkout-service or https://api.upstream.internal/v1"
              required
              className="w-full bg-canvas-subtleDark border border-white/10 rounded-lg px-3 py-1.5 text-xs text-white font-mono focus:border-brand-orange focus:outline-none"
            />
          </div>

          <div>
            <label className="block text-xs font-mono text-slate-400 mb-1">
              Required Schema Fields &amp; Validation Types (JSON format):
            </label>
            <textarea
              rows={6}
              value={fieldsJson}
              onChange={(e) => setFieldsJson(e.target.value)}
              className="w-full bg-[#0a0a0c] border border-white/10 rounded-lg p-3 text-xs font-mono text-emerald-400 focus:outline-none focus:border-brand-orange resize-none"
              spellCheck={false}
            />
          </div>

          {error && (
            <p className="text-xs text-rose-400 font-mono">
              {error}
            </p>
          )}

          <div className="pt-3 border-t border-white/10 flex items-center justify-end gap-2.5">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 rounded-lg text-xs font-mono text-slate-400 hover:text-white"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={isSaving}
              className="flex items-center gap-1.5 px-4 py-2 rounded-lg text-xs font-mono font-bold bg-brand-orange text-slate-950 hover:brightness-110 cursor-pointer disabled:opacity-50"
            >
              <Save className="w-3.5 h-3.5" />
              {isSaving ? 'Saving...' : 'Save & Upsert Contract'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
