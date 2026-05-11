export enum Kind {
  AgentRuntime = 'AgentRuntime',
  CodeInterpreter = 'CodeInterpreter',
}

export interface SandboxInfo {
  SessionID: string
  SandboxID: string
  Name: string
  SandboxNamespace: string
  Kind: Kind
  Image: string
  Status: string
  CreatedAt: string
  ExpiresAt: string
  Endpoints: Record<string, string>
}

export interface ListSandboxesResponse {
  total: number
  items: SandboxInfo[]
}

export interface AccessEndpoint {
  name: string
  url: string
}
