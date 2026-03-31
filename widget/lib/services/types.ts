export type PageResult<T> = {
  results: T[];
  page?: {
    page: number;
    limit: number;
    total: number;
  };
};

export type CursorResult<T> = {
  results: T[];
  cursor: string;
  hasMore: boolean;
};

export type JsonResult<T> = {
  success?: boolean;
  errorCode?: number;
  code?: number;
  message?: string;
  data?: T;
};

export type WidgetConversation = {
  id: number;
  subject?: string;
  status?: number;
  serviceMode?: number;
  currentAssigneeId?: number;
  lastMessageAt?: string;
  lastMessageSummary?: string;
  customerUnreadCount?: number;
  agentUnreadCount?: number;
  customerLastReadMessageId?: number;
  customerLastReadSeqNo?: number;
  customerLastReadAt?: string;
  agentLastReadMessageId?: number;
  agentLastReadSeqNo?: number;
  agentLastReadAt?: string;
};

export type WidgetMessage = {
  id: number;
  conversationId: number;
  senderType: string;
  senderName?: string;
  senderAvatar?: string;
  messageType: string;
  content: string;
  payload?: string;
  seqNo?: number;
  sentAt?: string;
  customerRead?: boolean;
  customerReadAt?: string;
  agentRead?: boolean;
  agentReadAt?: string;
};

export type WidgetAsset = {
  id: number;
  assetId: string;
  provider: string;
  filename: string;
  fileSize: number;
  mimeType: string;
  status: number;
  url: string;
  createdAt: string;
  updatedAt: string;
  createUserId: number;
  createUserName: string;
  updateUserId: number;
  updateUserName: string;
};

export type WidgetConfigResponse = {
  title?: string;
  subtitle?: string;
  welcomeText?: string;
  themeColor?: string;
};
