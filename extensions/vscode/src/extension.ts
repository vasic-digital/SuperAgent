import * as vscode from 'vscode';
import { HelixAgentClient } from './client';
import { ChatViewProvider } from './chatView';

let client: HelixAgentClient;

export function activate(context: vscode.ExtensionContext) {
    console.log('HelixAgent extension activated');

    // Initialize client
    const config = vscode.workspace.getConfiguration('helixagent');
    client = new HelixAgentClient({
        endpoint: config.get('endpoint') || 'http://localhost:7061',
        apiKey: config.get('apiKey') || '',
    });

    // Register chat view
    const chatProvider = new ChatViewProvider(context.extensionUri, client);
    context.subscriptions.push(
        vscode.window.registerWebviewViewProvider('helixagent.chat', chatProvider)
    );

    // Register commands
    context.subscriptions.push(
        vscode.commands.registerCommand('helixagent.startChat', () => {
            vscode.commands.executeCommand('workbench.view.extension.helixagent');
        }),

        vscode.commands.registerCommand('helixagent.explainCode', async () => {
            const editor = vscode.window.activeTextEditor;
            if (!editor) return;

            const selection = editor.document.getText(editor.selection);
            if (!selection) {
                vscode.window.showWarningMessage('Please select code to explain');
                return;
            }

            const explanation = await client.explainCode(selection, editor.document.languageId);
            chatProvider.showMessage(explanation);
        }),

        vscode.commands.registerCommand('helixagent.generateTests', async () => {
            const editor = vscode.window.activeTextEditor;
            if (!editor) return;

            const code = editor.document.getText(editor.selection) || editor.document.getText();
            const tests = await client.generateTests(code, editor.document.languageId);
            
            const testDoc = await vscode.workspace.openTextDocument({
                content: tests,
                language: editor.document.languageId,
            });
            await vscode.window.showTextDocument(testDoc);
        }),

        vscode.commands.registerCommand('helixagent.refactorCode', async () => {
            const editor = vscode.window.activeTextEditor;
            if (!editor) return;

            const selection = editor.selection;
            const code = editor.document.getText(selection);
            if (!code) {
                vscode.window.showWarningMessage('Please select code to refactor');
                return;
            }

            const refactored = await client.refactorCode(code, editor.document.languageId);
            
            const edit = new vscode.WorkspaceEdit();
            edit.replace(editor.document.uri, selection, refactored);
            await vscode.workspace.applyEdit(edit);
        })
    );
}

export function deactivate() {
    console.log('HelixAgent extension deactivated');
}
