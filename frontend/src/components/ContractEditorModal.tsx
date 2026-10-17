import React, { useState } from 'react'
import { ServiceContract } from '../types'
import { X, Layers, Plus, Save, Globe } from 'lucide-react'

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
    setContractId('custom-api')
    setContractName('Custom Microservice')
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
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 dark:bg-black/80 backdrop-blur-md animate-fade-in">
      <div className="w-full max-w-2xl bg-white dark:bg-canvas-cardDark border border-slate-200 dark:border-slate-800 rounded-3xl p-7 shadow-2xl flex flex-col max-h-[90vh]">
        <div className="flex items-center justify-between pb-4 border-b border-slate-200 dark:border-slate-800">
          <div className="flex items-center gap-2.5">
            <div className="p-2 rounded-xl bg-blue-50 dark:bg-blue-950/40 border border-blue-200 dark:border-blue-800/40 text-blue-600 dark:text-blue-400">
              <Layers className="w-5 h-5" />
            </div>
            <div>
              <h3 className="text-base font-bold text-slate-900 dark:text-white">
                Target Schema Contracts
              </h3>
              <p className="text-xs text-slate-500 dark:text-slate-400">
                Configure endpoints, required fields, and types
              </p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-1.5 rounded-xl text-slate-400 hover:text-slate-700 dark:hover:text-white hover:bg-slate-100 dark:hover:bg-slate-800 transition cursor-pointer"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Existing Contracts Selector & New button */}
        <div className="flex items-center gap-2.5 my-4">
          <select
            value={selectedId}
            onChange={(e) => setSelectedId(e.target.value)}
            className="flex-1 bg-slate-50 dark:bg-canvas-subtleDark border border-slate-200 dark:border-slate-800 rounded-xl px-3.5 py-2 text-xs font-medium text-slate-800 dark:text-slate-200 focus:outline-none focus:ring-2 focus:ring-brand-orange/40"
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
            className="flex items-center gap-1.5 px-3.5 py-2 rounded-xl text-xs font-semibold bg-slate-100 dark:bg-white/10 text-slate-800 dark:text-white hover:bg-slate-200 dark:hover:bg-white/20 transition cursor-pointer"
          >
            <Plus className="w-4 h-4" />
            <span>New Contract</span>
          </button>
        </div>

        {/* Form */}
        <form onSubmit={handleSave} className="space-y-4 overflow-auto flex-1 pr-1">
          <div className="grid grid-cols-2 gap-3.5">
            <div>
              <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">
                Contract ID (URL Slug):
              </label>
              <input
                type="text"
                value={contractId}
                onChange={(e) => setContractId(e.target.value)}
                required
                className="w-full bg-slate-50 dark:bg-canvas-subtleDark border border-slate-200 dark:border-slate-800 rounded-xl px-3.5 py-2 text-xs text-slate-800 dark:text-slate-200 font-mono focus:ring-2 focus:ring-brand-orange/40 focus:outline-none"
              />
            </div>
            <div>
              <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">
                Service Display Name:
              </label>
              <input
                type="text"
                value={contractName}
                onChange={(e) => setContractName(e.target.value)}
                required
                className="w-full bg-slate-50 dark:bg-canvas-subtleDark border border-slate-200 dark:border-slate-800 rounded-xl px-3.5 py-2 text-xs text-slate-800 dark:text-slate-200 font-medium focus:ring-2 focus:ring-brand-orange/40 focus:outline-none"
              />
            </div>
          </div>

          <div>
            <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1 flex items-center gap-1.5">
              <Globe className="w-3.5 h-3.5 text-cyan-500" />
              <span>Target Microservice Upstream Endpoint:</span>
            </label>
            <input
              type="text"
              value={endpoint}
              onChange={(e) => setEndpoint(e.target.value)}
              placeholder="e.g. internal://checkout-service or https://httpbin.org/post"
              required
              className="w-full bg-slate-50 dark:bg-canvas-subtleDark border border-slate-200 dark:border-slate-800 rounded-xl px-3.5 py-2 text-xs text-slate-800 dark:text-slate-200 font-mono focus:ring-2 focus:ring-brand-orange/40 focus:outline-none"
            />
          </div>

          <div>
            <label className="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-1">
              Required Schema Fields &amp; Validation Types (JSON):
            </label>
            <textarea
              rows={5}
              value={fieldsJson}
              onChange={(e) => setFieldsJson(e.target.value)}
              className="w-full bg-slate-900 border border-slate-800 rounded-xl p-3.5 text-xs font-mono text-emerald-400 focus:outline-none focus:ring-2 focus:ring-brand-orange/40 resize-none"
              spellCheck={false}
            />
          </div>

          {error && (
            <p className="text-xs text-rose-500 font-medium">
              {error}
            </p>
          )}

          <div className="pt-3 border-t border-slate-200 dark:border-slate-800 flex items-center justify-end gap-2.5">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 rounded-xl text-xs font-semibold text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={isSaving}
              className="flex items-center gap-1.5 px-5 py-2 rounded-xl text-xs font-bold bg-brand-orange text-slate-950 hover:brightness-105 cursor-pointer disabled:opacity-50 shadow-md"
            >
              <Save className="w-3.5 h-3.5" />
              <span>{isSaving ? 'Saving...' : 'Save & Upsert Contract'}</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
