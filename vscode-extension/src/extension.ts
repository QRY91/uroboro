import * as vscode from 'vscode';
import { exec } from 'child_process';
import { promisify } from 'util';
import * as path from 'path';

const execAsync = promisify(exec);

interface UroboroConfig {
    pythonPath: string;
    autoGitCapture: boolean;
    showStatusBar: boolean;
    captureTemplate: string;
}

class UroboroExtension {
    private statusBarItem: vscode.StatusBarItem;
    private config: UroboroConfig;
    private outputChannel: vscode.OutputChannel;

    constructor(context: vscode.ExtensionContext) {
        this.outputChannel = vscode.window.createOutputChannel('uroboro');
        this.statusBarItem = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Right, 100);
        this.updateConfig();
        this.setupStatusBar();
        this.setupGitIntegration();
    }

    private updateConfig() {
        const config = vscode.workspace.getConfiguration('uroboro');
        this.config = {
            pythonPath: config.get('pythonPath', 'python'),
            autoGitCapture: config.get('autoGitCapture', true),
            showStatusBar: config.get('showStatusBar', true),
            captureTemplate: config.get('captureTemplate', '{insight} (in {file}:{line})')
        };
    }

    private setupStatusBar() {
        if (this.config.showStatusBar) {
            this.statusBarItem.text = "$(notebook) uroboro";
            this.statusBarItem.tooltip = "Click to view uroboro status";
            this.statusBarItem.command = 'uroboro.status';
            this.statusBarItem.show();
            this.updateStatusBar();
        }
    }

    private async updateStatusBar() {
        try {
            const result = await this.runUroboroCommand('status --days 1');
            // Parse recent activity count from status output
            const activityMatch = result.match(/Recent activity \(\d+ days\): (\d+) items/);
            if (activityMatch) {
                const count = activityMatch[1];
                this.statusBarItem.text = `$(notebook) uroboro: ${count} recent`;
            }
        } catch (error) {
            this.statusBarItem.text = "$(warning) uroboro: error";
        }
    }

    private setupGitIntegration() {
        if (this.config.autoGitCapture) {
            // Listen for git operations (simplified - real implementation would use git extension API)
            const gitExtension = vscode.extensions.getExtension('vscode.git');
            if (gitExtension) {
                // TODO: Integrate with git extension for commit hooks
                this.log('Git integration initialized');
            }
        }
    }

    private async runUroboroCommand(command: string): Promise<string> {
        const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
        const cwd = workspaceFolder ? workspaceFolder.uri.fsPath : process.cwd();
        
        try {
            const { stdout, stderr } = await execAsync(
                `${this.config.pythonPath} -m src.cli ${command}`,
                { cwd }
            );
            
            if (stderr) {
                this.log(`Warning: ${stderr}`);
            }
            
            return stdout;
        } catch (error: any) {
            this.log(`Error running uroboro: ${error.message}`);
            throw error;
        }
    }

    // Command implementations
    async capture() {
        const insight = await vscode.window.showInputBox({
            prompt: 'What insight did you just discover?',
            placeholder: 'Fixed auth timeout - cut query time from 3s to 200ms',
            valueSelection: [0, 0]
        });

        if (!insight) return;

        try {
            await this.runUroboroCommand(`capture "${insight}"`);
            vscode.window.showInformationMessage(`✅ Captured: ${insight.substring(0, 50)}...`);
            this.updateStatusBar();
        } catch (error) {
            vscode.window.showErrorMessage('Failed to capture insight');
        }
    }

    async captureSelection() {
        const editor = vscode.window.activeTextEditor;
        if (!editor) return;

        const selection = editor.selection;
        const selectedText = editor.document.getText(selection);
        const fileName = path.basename(editor.document.fileName);
        const lineNumber = selection.start.line + 1;

        const insight = await vscode.window.showInputBox({
            prompt: 'Describe your insight about this code:',
            placeholder: 'This pattern solves the race condition we were seeing',
            valueSelection: [0, 0]
        });

        if (!insight) return;

        // Format with context using template
        const contextualInsight = this.config.captureTemplate
            .replace('{insight}', insight)
            .replace('{file}', fileName)
            .replace('{line}', lineNumber.toString());

        try {
            const command = `capture "${contextualInsight}" --tags code-context`;
            await this.runUroboroCommand(command);
            
            vscode.window.showInformationMessage(`✅ Captured with context: ${fileName}:${lineNumber}`);
            this.updateStatusBar();
        } catch (error) {
            vscode.window.showErrorMessage('Failed to capture insight with context');
        }
    }

    async showStatus() {
        try {
            const result = await this.runUroboroCommand('status --recent');
            
            // Create a new document to show the status
            const doc = await vscode.workspace.openTextDocument({
                content: result,
                language: 'plaintext'
            });
            
            await vscode.window.showTextDocument(doc);
        } catch (error) {
            vscode.window.showErrorMessage('Failed to get uroboro status');
        }
    }

    async publish() {
        const options = ['Blog Post', 'Social Content', 'Dev Log'];
        const choice = await vscode.window.showQuickPick(options, {
            placeHolder: 'What type of content do you want to publish?'
        });

        if (!choice) return;

        const typeMap: {[key: string]: string} = {
            'Blog Post': 'blog',
            'Social Content': 'social', 
            'Dev Log': 'devlog'
        };

        try {
            vscode.window.showInformationMessage('🔄 Generating content...');
            const result = await this.runUroboroCommand(`publish --type ${typeMap[choice]} --format markdown`);
            
            if (typeMap[choice] === 'blog' || typeMap[choice] === 'devlog') {
                // For markdown content, try to open the generated file
                const pathMatch = result.match(/saved to: (.+)/);
                if (pathMatch) {
                    const filePath = pathMatch[1].trim();
                    try {
                        const uri = vscode.Uri.file(filePath);
                        const doc = await vscode.workspace.openTextDocument(uri);
                        await vscode.window.showTextDocument(doc);
                        vscode.window.showInformationMessage(`✅ ${choice} opened for editing: ${path.basename(filePath)}`);
                    } catch (fileError) {
                        this.log(`Could not open file: ${filePath}, error: ${fileError}`);
                        vscode.window.showWarningMessage(`Content generated but could not open file: ${filePath}`);
                    }
                } else {
                    // If no file path found, create a new untitled document with the content
                    const doc = await vscode.workspace.openTextDocument({
                        content: result,
                        language: 'markdown'
                    });
                    await vscode.window.showTextDocument(doc);
                    vscode.window.showInformationMessage(`✅ ${choice} generated and opened for editing`);
                }
            } else {
                // For social content, show in output channel
                this.outputChannel.clear();
                this.outputChannel.appendLine(`=== ${choice.toUpperCase()} CONTENT ===`);
                this.outputChannel.appendLine(result);
                this.outputChannel.appendLine(`=== END ${choice.toUpperCase()} ===`);
                this.outputChannel.show();
                vscode.window.showInformationMessage(`✅ ${choice} generated - check output panel`);
            }
            
        } catch (error: any) {
            this.log(`Publish error: ${error.message}`);
            vscode.window.showErrorMessage(`Failed to publish ${choice.toLowerCase()}: ${error.message}`);
        }
    }

    async quickPublish() {
        try {
            vscode.window.showInformationMessage('🔄 Quick publishing blog post...');
            const result = await this.runUroboroCommand('publish --type blog --format markdown');
            
            // Try to open the generated markdown file
            const pathMatch = result.match(/saved to: (.+)/);
            if (pathMatch) {
                const filePath = pathMatch[1].trim();
                try {
                    const uri = vscode.Uri.file(filePath);
                    const doc = await vscode.workspace.openTextDocument(uri);
                    await vscode.window.showTextDocument(doc);
                    vscode.window.showInformationMessage(`✅ Blog post ready for editing: ${path.basename(filePath)}`);
                } catch (fileError) {
                    this.log(`Could not open generated file: ${filePath}`);
                    // Fallback: create new document with content
                    const doc = await vscode.workspace.openTextDocument({
                        content: result,
                        language: 'markdown'
                    });
                    await vscode.window.showTextDocument(doc);
                    vscode.window.showInformationMessage('✅ Blog post generated and opened for editing');
                }
            } else {
                // No file path found, show content in new document
                const doc = await vscode.workspace.openTextDocument({
                    content: result,
                    language: 'markdown'
                });
                await vscode.window.showTextDocument(doc);
                vscode.window.showInformationMessage('✅ Blog post generated and opened for editing');
            }
        } catch (error: any) {
            this.log(`Quick publish error: ${error.message}`);
            vscode.window.showErrorMessage(`Failed to quick publish: ${error.message}`);
        }
    }

    private log(message: string) {
        this.outputChannel.appendLine(`[${new Date().toISOString()}] ${message}`);
    }

    dispose() {
        this.statusBarItem.dispose();
        this.outputChannel.dispose();
    }
}

export function activate(context: vscode.ExtensionContext) {
    const uroboro = new UroboroExtension(context);

    // Register commands
    context.subscriptions.push(
        vscode.commands.registerCommand('uroboro.capture', () => uroboro.capture()),
        vscode.commands.registerCommand('uroboro.captureSelection', () => uroboro.captureSelection()),
        vscode.commands.registerCommand('uroboro.status', () => uroboro.showStatus()),
        vscode.commands.registerCommand('uroboro.publish', () => uroboro.publish()),
        vscode.commands.registerCommand('uroboro.quickPublish', () => uroboro.quickPublish())
    );

    // Register disposal
    context.subscriptions.push(uroboro);

    console.log('uroboro extension is now active! 🐍');
}

export function deactivate() {
    console.log('uroboro extension deactivated');
} 