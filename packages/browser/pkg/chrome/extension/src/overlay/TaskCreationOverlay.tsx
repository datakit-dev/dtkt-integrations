import { useState, useEffect } from 'react';
import { createPortal } from 'react-dom';
import type { CreateTaskRequest, ListSchemasRequest, CreateTaskResponse, ListSchemasResponse } from '../lib/messages';
import { fromJsonString } from '@bufbuild/protobuf';
import { ListExtractionSchemasResponseSchema, type ExtractionSchema } from '../proto/integration/browser/v1beta/extraction_pb';

interface TaskCreationOverlayProps {
  initialTitle: string;
  initialUrl: string;
  onClose: () => void;
}

export function TaskCreationOverlay({ initialTitle, initialUrl, onClose }: TaskCreationOverlayProps) {
  const [title, setTitle] = useState(initialTitle);
  const [url, setUrl] = useState(initialUrl);
  const [schemas, setSchemas] = useState<ExtractionSchema[]>([]);
  const [selectedSchemaId, setSelectedSchemaId] = useState('');
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Load available schemas on mount
  useEffect(() => {
    chrome.runtime.sendMessage({ type: 'LIST_SCHEMAS' } satisfies ListSchemasRequest)
      .then((response: ListSchemasResponse) => {
        if (response.schemasJson) {
          const resp = fromJsonString(ListExtractionSchemasResponseSchema, response.schemasJson);
          setSchemas(resp.schemas);
          if (resp.schemas.length > 0) {
            setSelectedSchemaId(resp.schemas[0].id);
          }
        }
        setLoading(false);
      })
      .catch(err => {
        console.error('Failed to load schemas:', err);
        setError('Failed to load extraction schemas');
        setLoading(false);
      });
  }, []);

  const handleCreate = async () => {
    if (!selectedSchemaId) {
      setError('Please select an extraction schema');
      return;
    }

    setCreating(true);
    setError(null);

    try {
      const response: CreateTaskResponse = await chrome.runtime.sendMessage({
        type: 'CREATE_TASK',
        title,
        url,
        schemaId: selectedSchemaId,
      } satisfies CreateTaskRequest);

      if (response.ok) {
        onClose();
      } else {
        setError(response.error || 'Failed to create task');
      }
    } catch (err) {
      console.error('Failed to create task:', err);
      setError(err instanceof Error ? err.message : 'Unknown error occurred');
    } finally {
      setCreating(false);
    }
  };

  return createPortal(
    <div style={styles.overlayRoot}>
      <div style={styles.backdrop} onClick={onClose} />
      <div style={styles.panel}>
        <div style={styles.header}>
          <h2 style={styles.title}>Create Task</h2>
          <button style={styles.closeButton} onClick={onClose} aria-label="Close">
            ×
          </button>
        </div>

        <div style={styles.content}>
          {loading ? (
            <div style={styles.loading}>Loading schemas...</div>
          ) : (
            <>
              <div style={styles.field}>
                <label style={styles.label} htmlFor="task-title">Title</label>
                <input
                  id="task-title"
                  type="text"
                  style={styles.input}
                  value={title}
                  onChange={(e) => setTitle(e.target.value)}
                  placeholder="Task title"
                />
              </div>

              <div style={styles.field}>
                <label style={styles.label} htmlFor="task-url">URL</label>
                <input
                  id="task-url"
                  type="url"
                  style={styles.input}
                  value={url}
                  onChange={(e) => setUrl(e.target.value)}
                  placeholder="https://..."
                />
              </div>

              <div style={styles.field}>
                <label style={styles.label} htmlFor="task-schema">Extraction Schema</label>
                <select
                  id="task-schema"
                  style={styles.select}
                  value={selectedSchemaId}
                  onChange={(e) => setSelectedSchemaId(e.target.value)}
                >
                  {schemas.length === 0 && (
                    <option value="">No schemas available</option>
                  )}
                  {schemas.map((schema) => (
                    <option key={schema.id} value={schema.id}>
                      {schema.name}
                    </option>
                  ))}
                </select>
              </div>

              {error && (
                <div style={styles.error}>{error}</div>
              )}

              <div style={styles.actions}>
                <button style={styles.cancelButton} onClick={onClose}>
                  Cancel
                </button>
                <button
                  style={{ ...styles.createButton, ...(creating ? styles.createButtonDisabled : {}) }}
                  onClick={handleCreate}
                  disabled={creating || !selectedSchemaId}
                >
                  {creating ? 'Creating...' : 'Create Task'}
                </button>
              </div>
            </>
          )}
        </div>
      </div>
    </div>,
    document.body
  );
}

const styles = {
  overlayRoot: {
    all: 'initial' as const,
    position: 'fixed' as const,
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    zIndex: 2147483647,
    fontFamily: 'system-ui, -apple-system, sans-serif',
    fontSize: '14px',
    pointerEvents: 'auto' as const,
  },
  backdrop: {
    position: 'fixed' as const,
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    backgroundColor: 'rgba(0, 0, 0, 0.5)',
  },
  panel: {
    position: 'fixed' as const,
    top: '50%',
    left: '50%',
    transform: 'translate(-50%, -50%)',
    backgroundColor: '#ffffff',
    borderRadius: '8px',
    boxShadow: '0 4px 24px rgba(0, 0, 0, 0.15)',
    width: '90%',
    maxWidth: '500px',
    maxHeight: '90vh',
    overflow: 'auto',
  },
  header: {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    padding: '20px 24px',
    borderBottom: '1px solid #e5e7eb',
  },
  title: {
    all: 'initial' as const,
    fontFamily: 'inherit',
    fontSize: '20px',
    fontWeight: '600',
    color: '#111827',
    margin: 0,
  },
  closeButton: {
    all: 'initial' as const,
    fontFamily: 'inherit',
    fontSize: '28px',
    fontWeight: '300',
    color: '#6b7280',
    background: 'none',
    border: 'none',
    cursor: 'pointer',
    padding: '0 8px',
    lineHeight: '1',
  },
  content: {
    padding: '24px',
  },
  loading: {
    textAlign: 'center' as const,
    padding: '40px',
    color: '#6b7280',
  },
  field: {
    marginBottom: '20px',
  },
  label: {
    all: 'initial' as const,
    fontFamily: 'inherit',
    display: 'block',
    fontSize: '14px',
    fontWeight: '500',
    color: '#374151',
    marginBottom: '8px',
  },
  input: {
    all: 'initial' as const,
    fontFamily: 'inherit',
    display: 'block',
    width: '100%',
    padding: '10px 12px',
    fontSize: '14px',
    border: '1px solid #d1d5db',
    borderRadius: '6px',
    boxSizing: 'border-box' as const,
    color: '#111827',
  },
  select: {
    all: 'initial' as const,
    fontFamily: 'inherit',
    display: 'block',
    width: '100%',
    padding: '10px 12px',
    fontSize: '14px',
    border: '1px solid #d1d5db',
    borderRadius: '6px',
    boxSizing: 'border-box' as const,
    color: '#111827',
    backgroundColor: '#ffffff',
    cursor: 'pointer',
  },
  error: {
    padding: '12px',
    backgroundColor: '#fef2f2',
    border: '1px solid #fecaca',
    borderRadius: '6px',
    color: '#991b1b',
    fontSize: '14px',
    marginBottom: '20px',
  },
  actions: {
    display: 'flex',
    gap: '12px',
    justifyContent: 'flex-end',
  },
  cancelButton: {
    all: 'initial' as const,
    fontFamily: 'inherit',
    padding: '10px 20px',
    fontSize: '14px',
    fontWeight: '500',
    color: '#374151',
    backgroundColor: '#ffffff',
    border: '1px solid #d1d5db',
    borderRadius: '6px',
    cursor: 'pointer',
  },
  createButton: {
    all: 'initial' as const,
    fontFamily: 'inherit',
    padding: '10px 20px',
    fontSize: '14px',
    fontWeight: '500',
    color: '#ffffff',
    backgroundColor: '#2563eb',
    border: 'none',
    borderRadius: '6px',
    cursor: 'pointer',
  },
  createButtonDisabled: {
    opacity: 0.6,
    cursor: 'not-allowed',
  },
};
