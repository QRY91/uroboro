import * as vscode from 'vscode';
import { exec } from 'child_process';
import { promisify } from 'util';
import * as path from 'path';

const execAsync = promisify(exec);

export function activate(context: vscode.ExtensionContext) {
    console.log('uroboro extension activating...');
    
    // Super minimal - just register commands, do nothing else
    context.subscriptions.push(
        vscode.commands.registerCommand('uroboro.capture', async () => {
            try {
                console.log('uroboro.capture called');
                
                const insight = await vscode.window.showInputBox({
                    prompt: '💡 Enter your development insight:',
                    placeHolder: 'Fixed auth timeout - cut query time from 3s to 200ms'
                });

                if (!insight?.trim()) {
                    vscode.window.showWarningMessage('No insight entered');
                    return;
                }

                // Find binary
                const binaryPaths = [
                    '/home/qry/stuff/projects/uroboro/go/uroboro',
                    path.join(__dirname, 'go-uroboro'),
                    '/home/qry/stuff/projects/uroboro/uroboro-go'
                ];
                
                let goBinaryPath = '';
                for (const binPath of binaryPaths) {
                    try {
                        if (require('fs').existsSync(binPath)) {
                            goBinaryPath = binPath;
                            break;
                        }
                    } catch (e) {
                        console.log(`Error checking ${binPath}:`, e);
                    }
                }
                
                if (!goBinaryPath) {
                    vscode.window.showErrorMessage('uroboro binary not found');
                    return;
                }
                
                console.log(`Using binary: ${goBinaryPath}`);
                
                // Run command with timeout
                const command = `${goBinaryPath} capture "${insight.trim()}"`;
                console.log(`Running: ${command}`);
                
                const { stdout, stderr } = await execAsync(command, { timeout: 10000 });
                
                if (stderr) {
                    console.log('stderr:', stderr);
                }
                
                console.log('stdout:', stdout);
                vscode.window.showInformationMessage(`✅ Captured: ${insight.substring(0, 50)}...`);
                
            } catch (error: any) {
                console.error('Capture error:', error);
                vscode.window.showErrorMessage(`Capture failed: ${error.message}`);
            }
        }),
        
        vscode.commands.registerCommand('uroboro.status', async () => {
            try {
                console.log('uroboro.status called');
                vscode.window.showInformationMessage('Status command - simplified for debugging');
            } catch (error: any) {
                console.error('Status error:', error);
                vscode.window.showErrorMessage(`Status failed: ${error.message}`);
            }
        })
    );

    console.log('uroboro extension activated successfully');
}

export function deactivate() {
    console.log('uroboro extension deactivated');
} 